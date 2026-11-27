// Pembaca daftar file ZIP di browser. Hanya membaca central directory,
// tidak mengekstrak atau menjalankan isi arsip. Sebagai tambahan, sejumlah
// kecil file teks dibaca untuk mendeteksi import framework yang tidak
// terlihat dari nama file saja (mis. bundle hasil build).

const EOCD_SIGNATURE = 0x06054b50;
const ZIP64_LOCATOR_SIGNATURE = 0x07064b50;
const ZIP64_EOCD_SIGNATURE = 0x06064b50;
const CENTRAL_ENTRY_SIGNATURE = 0x02014b50;
const LOCAL_ENTRY_SIGNATURE = 0x04034b50;
const MAX_PREVIEW_FILES = 5000;
const MAX_ZIP_BYTES = 100 * 1024 * 1024;

const SCAN_EXT = /\.(?:html?|js|mjs|cjs|jsx|ts|tsx|vue|svelte)$/i;
const MAX_SCAN_FILES = 40;
const MAX_SCAN_BYTES = 160 * 1024;
const MAX_TOTAL_SCAN_BYTES = 2 * 1024 * 1024;
const MAX_CONCURRENT_SCANS = 2;

let scanGeneration = 0;
let activeScans = 0;
const scanQueue = [];

const FRAMEWORK_CONTENT = [
  /(?:from\s*|require\(\s*|import\(\s*)['"](?:react|react-dom|next(?:\/[\w.-]+)?|vue|nuxt|svelte|@sveltejs\/kit|@angular\/core|astro|preact|solid-js|ember-source|vite)['"]/i,
  /<script[^>]+src=['"][^'"]*(?:react|vue|svelte|angular|next)[^'"]*['"]/i,
  /\b(?:ReactDOM\.createRoot|createRoot\(|new Vue\(|Vue\.createApp|SvelteComponent|_next\/static)/i,
];

function superseded() {
  const error = new Error('ZIP scan superseded by a newer selection');
  error.name = 'AbortError';
  return error;
}

function ensureCurrent(token) {
  if (token !== scanGeneration) throw superseded();
}

function pumpScans() {
  while (activeScans < MAX_CONCURRENT_SCANS && scanQueue.length) {
    const job = scanQueue.shift();
    if (job.token !== scanGeneration) {
      job.reject(superseded());
      continue;
    }
    activeScans += 1;
    Promise.resolve()
      .then(job.task)
      .then(job.resolve, job.reject)
      .finally(() => {
        activeScans -= 1;
        pumpScans();
      });
  }
}

// Satu antrean kecil mencegah pembacaan ZIP bertumpuk. Setiap seleksi baru
// menaikkan generation: pekerjaan lama yang masih antre/berjalan dibatalkan.
export function listZipPaths(file) {
  const token = ++scanGeneration;
  return new Promise((resolve, reject) => {
    scanQueue.push({ task: () => readZipPaths(file, token), token, resolve, reject });
    pumpScans();
  });
}

async function readZipPaths(file, token) {
  ensureCurrent(token);
  if (!file || !/\.zip$/i.test(file.name)) {
    throw new Error('Choose a .zip file.');
  }
  if (file.size > MAX_ZIP_BYTES) {
    throw new Error('This ZIP is larger than 100 MB.');
  }

  const buffer = await file.arrayBuffer();
  ensureCurrent(token);
  const view = new DataView(buffer);
  const bytes = new Uint8Array(buffer);
  const eocd = findEndRecord(view);
  if (eocd < 0) {
    throw new Error('Could not read this ZIP archive.');
  }

  let entries = view.getUint16(eocd + 10, true);
  let offset = view.getUint32(eocd + 16, true);

  // Arsip dengan banyak entri memakai ZIP64 record sebelum EOCD biasa.
  if (entries === 0xffff || offset === 0xffffffff) {
    const locator = eocd - 20;
    if (
      locator >= 0 &&
      view.byteLength >= locator + 20 &&
      view.getUint32(locator, true) === ZIP64_LOCATOR_SIGNATURE
    ) {
      const zip64 = Number(view.getBigUint64(locator + 8, true));
      if (
        zip64 >= 0 &&
        zip64 + 56 <= view.byteLength &&
        view.getUint32(zip64, true) === ZIP64_EOCD_SIGNATURE
      ) {
        entries = Number(view.getBigUint64(zip64 + 32, true));
        offset = Number(view.getBigUint64(zip64 + 48, true));
      }
    }
  }

  if (!Number.isSafeInteger(offset) || offset < 0 || offset >= view.byteLength) {
    throw new Error('Could not read the file list in this ZIP.');
  }

  const decoder = new TextDecoder('utf-8');
  const paths = [];
  const scanCandidates = [];
  let cursor = offset;
  let truncated = false;

  for (let i = 0; i < entries && cursor + 46 <= view.byteLength; i++) {
    if (i % 128 === 0) ensureCurrent(token);
    if (view.getUint32(cursor, true) !== CENTRAL_ENTRY_SIGNATURE) break;

    const method = view.getUint16(cursor + 10, true);
    const compressedSize = view.getUint32(cursor + 20, true);
    const uncompressedSize = view.getUint32(cursor + 24, true);
    const nameLength = view.getUint16(cursor + 28, true);
    const extraLength = view.getUint16(cursor + 30, true);
    const commentLength = view.getUint16(cursor + 32, true);
    const localOffset = view.getUint32(cursor + 42, true);
    const nameStart = cursor + 46;
    if (nameStart + nameLength > view.byteLength) break;

    const raw = decoder.decode(bytes.subarray(nameStart, nameStart + nameLength));
    cursor = nameStart + nameLength + extraLength + commentLength;

    if (raw.endsWith('/')) continue;
    const path = cleanName(raw);
    if (!path) continue;
    if (paths.length >= MAX_PREVIEW_FILES) {
      truncated = true;
      continue;
    }
    paths.push(path);
    if (SCAN_EXT.test(path) && uncompressedSize <= MAX_SCAN_BYTES) {
      scanCandidates.push({ path, method, compressedSize, uncompressedSize, localOffset });
    }
  }

  if (paths.length === 0 && !truncated) {
    throw new Error('No files were found in this ZIP.');
  }

  ensureCurrent(token);
  const signals = await scanForFrameworkSignals(view, bytes, scanCandidates, decoder, token);
  ensureCurrent(token);
  return { paths, truncated, signals };
}

function findEndRecord(view) {
  if (view.byteLength < 22) return -1;
  const minimum = Math.max(0, view.byteLength - 65557);
  for (let i = view.byteLength - 22; i >= minimum; i--) {
    if (view.getUint32(i, true) === EOCD_SIGNATURE) return i;
  }
  return -1;
}

function cleanName(raw) {
  let path = String(raw).replace(/\\/g, '/').replace(/^\.\//, '').trim();
  if (!path || path.startsWith('/') || /^[a-z]:/i.test(path)) return '';
  if (path.split('/').some((part) => !part || part === '.' || part === '..')) return '';
  return path;
}

// Membaca maksimal 40 file teks kecil dari arsip untuk mencari petunjuk
// framework (import react/vue/svelte/dll). Gagal baca satu entri diabaikan;
// pemeriksaan berdasarkan nama file tetap berjalan.
async function scanForFrameworkSignals(view, bytes, candidates, decoder, token) {
  if (typeof DecompressionStream === 'undefined') return [];

  const signals = [];
  let total = 0;
  for (const item of candidates.slice(0, MAX_SCAN_FILES)) {
    ensureCurrent(token);
    if (total >= MAX_TOTAL_SCAN_BYTES) break;
    const local = item.localOffset;
    if (!Number.isSafeInteger(local) || local < 0 || local + 30 > view.byteLength) continue;
    if (view.getUint32(local, true) !== LOCAL_ENTRY_SIGNATURE) continue;

    const nameLength = view.getUint16(local + 26, true);
    const extraLength = view.getUint16(local + 28, true);
    const start = local + 30 + nameLength + extraLength;
    const end = start + item.compressedSize;
    if (start >= end || end > view.byteLength || item.uncompressedSize > MAX_SCAN_BYTES) continue;

    try {
      let raw = bytes.subarray(start, end);
      if (item.method === 8) {
        const stream = new Blob([raw]).stream().pipeThrough(new DecompressionStream('deflate-raw'));
        raw = new Uint8Array(await new Response(stream).arrayBuffer());
      } else if (item.method !== 0) {
        continue;
      }
      ensureCurrent(token);
      if (raw.byteLength > MAX_SCAN_BYTES) continue;
      total += raw.byteLength;

      const text = decoder.decode(raw);
      for (const pattern of FRAMEWORK_CONTENT) {
        const match = text.match(pattern);
        if (match) {
          signals.push({ path: item.path, signal: match[0].slice(0, 80) });
          break;
        }
      }
    } catch (err) {
      if (err?.name === 'AbortError') throw err;
      // Entri yang tidak bisa dibuka tidak membuat proses gagal.
    }
  }
  return signals;
}
