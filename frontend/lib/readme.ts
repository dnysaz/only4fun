/**
 * README datang dari GitHub sebagai HTML. GitHub sudah membersihkannya, tetapi
 * kita tetap menjalankan allowlist sanitizer di browser sebagai lapisan kedua:
 * script/iframe/style, atribut `on*`, dan URL berbahaya dibuang sebelum HTML
 * dimasukkan lewat dangerouslySetInnerHTML.
 *
 * Fungsi ini butuh DOM (DOMParser). Saat SSR ia mengembalikan string kosong;
 * komponen pemanggil baru merender hasilnya setelah mount.
 */

// Elemen yang aman untuk konten README.
const ALLOWED_TAGS = new Set([
  "a", "abbr", "audio", "b", "blockquote", "br", "caption", "code", "dd",
  "del", "details", "dfn", "div", "dl", "dt", "em", "figcaption", "figure",
  "h1", "h2", "h3", "h4", "h5", "h6", "hr", "i", "img", "ins", "kbd", "li",
  "mark", "ol", "p", "picture", "pre", "q", "s", "samp", "small", "source",
  "span", "strong", "sub", "summary", "sup", "table", "tbody", "td", "tfoot",
  "th", "thead", "tr", "track", "u", "ul", "var", "video",
])

// Elemen yang dibuang beserta isinya (bukan sekadar di-unwrap).
const DROP_TAGS = new Set([
  "applet", "base", "button", "embed", "form", "frame", "frameset", "iframe",
  "input", "link", "math", "meta", "noscript", "object", "option", "script",
  "select", "style", "svg", "template", "textarea",
])

// Atribut yang boleh tersisa. `id`/`name` sengaja tidak ada supaya HTML tidak
// bisa menimpa variabel/elemen global (DOM clobbering).
const ALLOWED_ATTRS = new Set([
  "align", "alt", "checked", "class", "colspan", "controls", "dir", "disabled",
  "height", "href", "lang", "media", "open", "poster", "rel", "reversed",
  "rowspan", "src", "srcset", "start", "target", "title", "type", "width",
])

const SCHEME_PATTERN = /^[a-z][a-z0-9+.-]*:/i

/** Buang karakter kontrol yang bisa menyamarkan skema (mis. "java\nscript:"). */
function normalizeUrl(value: string): string {
  return value.replace(/[\u0000-\u001f\u007f]/g, "").trim()
}

function isSafeLink(value: string): boolean {
  const url = normalizeUrl(value).toLowerCase()
  if (!url) return false
  if (url.startsWith("#") || url.startsWith("/")) return true
  if (url.startsWith("http://") || url.startsWith("https://") || url.startsWith("mailto:")) {
    return true
  }
  // URL relatif tanpa skema (mis. "docs/page.md").
  return !SCHEME_PATTERN.test(url)
}

function isSafeResource(value: string): boolean {
  const url = normalizeUrl(value).toLowerCase()
  if (!url) return false
  if (url.startsWith("/")) return true
  if (url.startsWith("http://") || url.startsWith("https://")) return true
  return !SCHEME_PATTERN.test(url)
}

function cleanSrcset(value: string): string {
  const candidates = value
    .split(",")
    .map((candidate) => candidate.trim())
    .filter(Boolean)
    .filter((candidate) => isSafeResource(candidate.split(/\s+/)[0] || ""))
  return candidates.join(", ")
}

function cleanElement(el: Element, tag: string) {
  for (const attr of Array.from(el.attributes)) {
    const name = attr.name.toLowerCase()
    if (name.startsWith("on") || !ALLOWED_ATTRS.has(name)) {
      el.removeAttribute(attr.name)
      continue
    }
    if (name === "href" && !isSafeLink(attr.value)) {
      el.removeAttribute(attr.name)
      continue
    }
    if ((name === "src" || name === "poster") && !isSafeResource(attr.value)) {
      el.removeAttribute(attr.name)
      continue
    }
    if (name === "srcset") {
      const cleaned = cleanSrcset(attr.value)
      if (cleaned) el.setAttribute("srcset", cleaned)
      else el.removeAttribute(attr.name)
    }
  }

  // Link eksternal tidak boleh mendapat akses ke window pembuka.
  if (tag === "a" && el.hasAttribute("href")) {
    el.setAttribute("rel", "nofollow noopener noreferrer")
  }
}

function cleanChildren(parent: Node) {
  for (const child of Array.from(parent.childNodes)) {
    if (child.nodeType === 8) {
      // Komentar tidak diperlukan dan bisa menyembunyikan markup.
      child.remove()
      continue
    }
    if (child.nodeType !== 1) continue

    const el = child as Element
    const tag = el.tagName.toLowerCase()

    if (DROP_TAGS.has(tag)) {
      el.remove()
      continue
    }

    if (!ALLOWED_TAGS.has(tag)) {
      // Tag tak dikenal di-unwrap: isinya tetap tampil, elemennya hilang.
      cleanChildren(el)
      el.replaceWith(...Array.from(el.childNodes))
      continue
    }

    cleanElement(el, tag)
    cleanChildren(el)
  }
}

/** Membersihkan HTML README dan menambahkan atribut performa pada gambar. */
export function sanitizeReadmeHtml(html: string): string {
  if (!html) return ""
  if (typeof window === "undefined" || typeof DOMParser === "undefined") return ""

  const doc = new DOMParser().parseFromString(html, "text/html")
  cleanChildren(doc.body)
  return optimizeReadmeHtml(doc.body.innerHTML)
}

/**
 * README datang sebagai HTML yang sudah dirender GitHub. Gambarnya belum punya
 * atribut performa, jadi ditambahkan di sini: lazy loading, decoding async, dan
 * ukuran intrinsik agar layout tidak bergeser saat gambar muncul.
 */
export function optimizeReadmeHtml(html: string): string {
  return html.replace(/<img\s([^>]*)>/gi, (tag, attrs: string) => {
    let next = attrs
    if (!/\bloading=/i.test(next)) next += ' loading="lazy"'
    if (!/\bdecoding=/i.test(next)) next += ' decoding="async"'
    if (!/\breferrerpolicy=/i.test(next)) next += ' referrerpolicy="no-referrer"'
    if (/\bwidth=/i.test(next) && /\bheight=/i.test(next)) {
      next += ' style="height:auto"'
    }
    return `<img ${next.trim()}>`
  })
}
