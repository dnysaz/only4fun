// check-word.ts
// Validasi nama project (subdomain) dari sisi client.
// Melindungi form dari input berbahaya (XSS, CSRF, SQL injection) dan
// menolak kata-kata yang tidak pantas / terlarang.

// ─── Daftar kata terlarang ────────────────────────────────────────────────

// Kata kasar (English + Indonesian)
const PROFANITY: string[] = [
  // English
  "fuck", "shit", "ass", "bitch", "damn", "crap", "dick", "piss", "cock", "pussy",
  "bastard", "whore", "slut", "fag", "nigger", "retard", "cunt", "twat", "wank",
  "bollocks", "arse", "prick", "douche", "motherfucker", "bullshit", "asshole",
  "jackass", "dumbass", "dipshit", "shithead", "fuckface", "fucktard",
  // Indonesian
  "anjing", "bangsat", "babi", "kontol", "memek", "ngentot", "jembut", "tolol",
  "goblok", "idiot", "bodoh", "sialan", "kampret", "brengsek", "keparat",
  "asu", "monyet", "tai", "perek", "lonte", "sundal", "jalang", "pelacur",
  "kampungan", "norak", "dungu", "bego", "sinting", "gila", "edan",
  "setan", "iblis", "laknat", "celaka", "sial", "bangke",
]

// Kata administratif / infrastruktur
const ADMINISTRATIVE: string[] = [
  "admin", "administrator", "root", "superuser", "sysadmin", "moderator",
  "mod", "staff", "operator", "manager", "owner", "support", "helpdesk",
  "system", "sys", "daemon", "service", "services", "backend", "frontend",
  "api", "www", "mail", "ftp", "smtp", "pop", "imap", "dns", "ns", "mx",
  "cdn", "proxy", "gateway", "router", "firewall", "server", "host", "localhost",
  "webmaster", "postmaster", "abuse", "security", "noc", "ops", "devops",
  "infrastructure", "infra", "deploy", "deployment", "build", "ci", "cd",
  "pipeline", "runner", "worker", "queue", "cache", "redis", "database", "db",
  "sql", "mysql", "postgres", "mongo", "redis", "elastic", "kafka",
  "nginx", "apache", "caddy", "traefik", "haproxy", "varnish",
  "ssl", "tls", "cert", "certificate", "key", "secret", "token", "auth",
  "oauth", "jwt", "session", "cookie", "cors", "csrf", "xss",
  "config", "configuration", "env", "environment", "settings", "setup",
  "install", "update", "upgrade", "patch", "release", "version", "v1", "v2",
  "beta", "alpha", "staging", "production", "prod", "dev", "development",
  "test", "testing", "qa", "uat", "sandbox", "demo", "example", "sample",
  "temp", "tmp", "backup", "restore", "archive", "log", "logs", "debug",
  "trace", "monitor", "monitoring", "metrics", "health", "status", "ping",
  "webhook", "callback", "hook", "trigger", "event", "events",
  "user", "users", "account", "accounts", "profile", "profiles",
  "project", "projects", "site", "sites", "page", "pages", "app", "apps",
  "dashboard", "panel", "console", "terminal", "shell", "command",
  "login", "logout", "register", "signup", "signin", "auth", "oauth",
  "password", "passwd", "credential", "credentials", "permission", "permissions",
  "role", "roles", "group", "groups", "team", "teams", "org", "organization",
  "billing", "payment", "invoice", "subscription", "plan", "pricing",
  "notification", "notifications", "alert", "alerts", "message", "messages",
  "email", "emails", "mail", "mailbox", "inbox", "outbox", "newsletter",
  "search", "find", "query", "filter", "sort", "order", "list", "create",
  "update", "delete", "remove", "add", "edit", "view", "show", "hide",
  "enable", "disable", "activate", "deactivate", "start", "stop", "restart",
  "run", "execute", "process", "job", "task", "schedule", "cron",
  "import", "export", "upload", "download", "file", "files", "folder",
  "directory", "path", "route", "routing", "redirect", "rewrite",
  "static", "assets", "public", "media", "image", "images", "video",
  "audio", "document", "documents", "attachment", "attachments",
  "api", "rest", "graphql", "grpc", "websocket", "ws", "sse",
  "json", "xml", "html", "css", "js", "javascript", "typescript",
  "node", "npm", "yarn", "pnpm", "go", "golang", "rust", "python",
  "java", "php", "ruby", "dotnet", "csharp", "swift", "kotlin",
  "docker", "kubernetes", "k8s", "helm", "terraform", "ansible",
  "aws", "azure", "gcp", "cloud", "cloudflare", "vercel", "netlify",
  "github", "gitlab", "bitbucket", "git", "svn", "mercurial",
  "slack", "discord", "telegram", "whatsapp", "twitter", "facebook",
  "instagram", "linkedin", "youtube", "tiktok", "reddit",
  "google", "microsoft", "apple", "amazon", "meta", "openai",
  "stripe", "paypal", "visa", "mastercard", "amex",
  "only4", "only4fun", "fun", "only",
]

// Kata pornografi
const PORNOGRAPHIC: string[] = [
  "porn", "porno", "pornography", "xxx", "sex", "sexy", "nude", "naked",
  "nsfw", "erotic", "erotica", "fetish", "bdsm", "hentai", "ecchi",
  "milf", "dildo", "vibrator", "orgasm", "masturbate", "masturbation",
  "blowjob", "handjob", "cumshot", "creampie", "gangbang", "threesome",
  "anal", "oral", "vaginal", "penis", "vagina", "breast", "boob", "tit",
  "ass", "butt", "booty", "thong", "lingerie", "stripper", "strip",
  "escort", "brothel", "prostitute", "hooker", "callgirl", "camgirl",
  "onlyfans", "fansly", "chaturbate", "pornhub", "xvideos", "xnxx",
  "redtube", "youporn", "tube8", "spankbang", "xhamster",
  "bokep", "porno", "mesum", "syur", "vulgar", "cabul", "asusila",
  "telanjang", "bugil", "hot", "sexy", "erotis", "birahi", "nafsu",
  "sange", "horny", "klimaks", "sperma", "mani", "vagina", "penis",
  "payudara", "pantat", "paha", "selangkangan", "zakar", "vulva",
  "klitoris", "testis", "skrotum", "anus", "rectum",
]

// Kata politik / nama tokoh publik
const POLITICAL: string[] = [
  "president", "presiden", "pm", "prime", "minister", "menteri",
  "governor", "gubernur", "mayor", "walikota", "regent", "bupati",
  "senator", "dpr", "mp", "parliament", "parlemen", "congress",
  "senate", "senat", "assembly", "majelis", "council", "dewan",
  "politician", "politisi", "politics", "politik", "party", "partai",
  "election", "pemilu", "vote", "voting", "campaign", "kampanye",
  "democrat", "republican", "conservative", "liberal", "socialist",
  "communist", "fascist", "nazi", "dictator", "dictatorship",
  "monarchy", "kerajaan", "sultan", "raja", "ratu", "prince", "princess",
  "jokowi", "prabowo", "megawati", "sby", "habibie", "soekarno",
  "soeharto", "gusdur", "jusuf", "anies", "ganjar", "ridwan",
  "trump", "biden", "obama", "clinton", "bush", "reagan", "kennedy",
  "putin", "xi", "jinping", "kim", "jong", "un", "modi", "macron",
  "merkel", "johnson", "sunak", "trudeau", "zelensky", "netanyahu",
  "erdogan", "bolsonaro", "lula", "sanchez", "scholz",
  "elon", "musk", "bezos", "zuckerberg", "gates", "jobs", "cook",
  "taylor", "swift", "beyonce", "rihanna", "drake", "kanye", "west",
  "bieber", "gaga", "madonna", "eminem", "jayz", "beyonce",
  "messi", "ronaldo", "neymar", "mbappe", "lebron", "jordan",
  "federer", "nadal", "serena", "williams", "tiger", "woods",
  "putin", "stalin", "hitler", "mao", "lenin", "trotsky",
  "binladen", "osama", "sadat", "gaddafi", "assad", "saddam",
  "manson", "bundy", "dahmer", "gacy", "ripper", "zodiac",
]

// Semua kata terlarang digabung
const FORBIDDEN_WORDS: string[] = [
  ...PROFANITY,
  ...ADMINISTRATIVE,
  ...PORNOGRAPHIC,
  ...POLITICAL,
]

// ─── Pola berbahaya (XSS, CSRF, SQL injection) ───────────────────────────

// Pola XSS
const XSS_PATTERNS: RegExp[] = [
  /<script/i,
  /<\/script/i,
  /javascript:/i,
  /on\w+\s*=/i,          // onclick=, onerror=, etc.
  /<iframe/i,
  /<object/i,
  /<embed/i,
  /<svg/i,
  /<img/i,
  /<a\s/i,
  /expression\(/i,
  /vbscript:/i,
  /data:/i,
  /&#/i,
  /\\x/i,
  /eval\(/i,
  /setTimeout/i,
  /setInterval/i,
  /document\./i,
  /window\./i,
  /alert\(/i,
  /prompt\(/i,
  /confirm\(/i,
]

// Pola SQL injection
const SQL_PATTERNS: RegExp[] = [
  /('|"|;|--|\/\*|\*\/)/,
  /\b(SELECT|INSERT|UPDATE|DELETE|DROP|CREATE|ALTER|TRUNCATE|UNION|JOIN|WHERE|FROM|TABLE|DATABASE)\b/i,
  /\b(OR|AND)\s+\d+\s*=\s*\d+/i,
  /\b(OR|AND)\s+'[^']*'\s*=\s*'[^']*'/i,
  /\bEXEC\b/i,
  /\bEXECUTE\b/i,
  /\bxp_/i,
  /\bsp_/i,
  /\bCAST\b/i,
  /\bCONVERT\b/i,
  /\bCHAR\b/i,
  /\bNCHAR\b/i,
  /\bVARCHAR\b/i,
  /\bDECLARE\b/i,
  /\bWAITFOR\b/i,
  /\bDELAY\b/i,
  /\bSHUTDOWN\b/i,
  /\bINFORMATION_SCHEMA\b/i,
  /\bSYSOBJECTS\b/i,
  /\bSYSCOLUMNS\b/i,
]

// Pola CSRF / token-like
const CSRF_PATTERNS: RegExp[] = [
  /csrf/i,
  /xsrf/i,
  /token/i,
  /session/i,
  /cookie/i,
  /auth/i,
  /bearer/i,
  /oauth/i,
  /jwt/i,
]

// ─── Fungsi validasi ──────────────────────────────────────────────────────

export interface ValidationResult {
  valid: boolean
  error: string
}

/**
 * Validasi nama project (subdomain).
 * Mengembalikan { valid: true } jika aman, atau { valid: false, error: "..." } jika tidak.
 */
export function validateProjectName(name: string): ValidationResult {
  const trimmed = name.trim()

  // 1. Minimal panjang
  if (trimmed.length < 3) {
    return { valid: false, error: "Project name must be at least 3 characters long." }
  }

  // 2. Maksimal panjang (DNS label max 63)
  if (trimmed.length > 63) {
    return { valid: false, error: "Project name must be at most 63 characters." }
  }

  // 3. Hanya lowercase, angka, dash
  if (!/^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$/.test(trimmed)) {
    return { valid: false, error: "Project name may only contain lowercase letters, numbers and dashes, and must not start or end with a dash." }
  }

  // 4. Tidak boleh hanya angka
  if (/^\d+$/.test(trimmed)) {
    return { valid: false, error: "Project name cannot be only numbers." }
  }

  // 5. Tidak boleh hanya dash
  if (/^-+$/.test(trimmed)) {
    return { valid: false, error: "Project name cannot be only dashes." }
  }

  // 6. Cek kata terlarang (case-insensitive, whole-word / token match)
  const lower = trimmed.toLowerCase()
  const tokens = lower.split("-")
  for (const word of FORBIDDEN_WORDS) {
    const w = word.toLowerCase()
    for (const token of tokens) {
      if (token === w) {
        return { valid: false, error: `Project name contains a restricted word: "${word}". Please choose another name.` }
      }
    }
  }

  // 7. Cek pola XSS
  for (const pattern of XSS_PATTERNS) {
    if (pattern.test(trimmed)) {
      return { valid: false, error: "Project name contains invalid characters (possible XSS)." }
    }
  }

  // 8. Cek pola SQL injection
  for (const pattern of SQL_PATTERNS) {
    if (pattern.test(trimmed)) {
      return { valid: false, error: "Project name contains invalid characters (possible SQL injection)." }
    }
  }

  // 9. Cek pola CSRF/token
  for (const pattern of CSRF_PATTERNS) {
    if (pattern.test(trimmed)) {
      return { valid: false, error: "Project name contains a restricted term." }
    }
  }

  // 10. Cek karakter berbahaya (selain a-z, 0-9, -)
  if (/[<>'"\\\/;(){}[\]|&$#@!%^*+=~`]/.test(trimmed)) {
    return { valid: false, error: "Project name contains special characters that are not allowed." }
  }

  // 11. Cek unicode/encoded characters
  if (/[^\x00-\x7F]/.test(trimmed)) {
    return { valid: false, error: "Project name must use ASCII characters only." }
  }

  // 12. Cek path traversal
  if (trimmed.includes("..") || trimmed.includes("./") || trimmed.includes("/")) {
    return { valid: false, error: "Project name contains invalid path characters." }
  }

  // 13. Cek reserved subdomains (sudah di-handle backend, tapi double check)
  const reserved = ["www", "api", "app", "admin", "auth", "docs", "help", "profile", "dashboard", "mail"]
  if (reserved.includes(trimmed)) {
    return { valid: false, error: `Project name "${trimmed}" is reserved. Please choose another.` }
  }

  return { valid: true, error: "" }
}

/**
 * Sanitize input untuk mencegah XSS.
 * Menghapus semua tag HTML dan karakter berbahaya.
 */
export function sanitizeInput(input: string): string {
  return input
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#x27;")
    .replace(/\//g, "&#x2F;")
    .replace(/`/g, "&#96;")
    .replace(/=/g, "&#61;")
}

/**
 * Validasi format env var key.
 */
export function validateEnvKey(key: string): ValidationResult {
  const trimmed = key.trim()
  if (trimmed.length === 0) {
    return { valid: false, error: "Environment variable name cannot be empty." }
  }
  if (trimmed.length > 256) {
    return { valid: false, error: "Environment variable name is too long." }
  }
  if (!/^[A-Za-z_][A-Za-z0-9_]*$/.test(trimmed)) {
    return { valid: false, error: "Environment variable name must start with a letter or underscore and contain only letters, numbers, and underscores." }
  }
  // Cek kata terlarang
  const lower = trimmed.toLowerCase()
  for (const word of FORBIDDEN_WORDS) {
    if (lower.includes(word.toLowerCase())) {
      return { valid: false, error: `Environment variable name contains a restricted word: "${word}".` }
    }
  }
  return { valid: true, error: "" }
}

/**
 * Validasi format env var value.
 */
export function validateEnvValue(value: string): ValidationResult {
  // Value boleh kosong (untuk hapus)
  if (value.length > 4096) {
    return { valid: false, error: "Environment variable value is too long (max 4096 characters)." }
  }
  // Cek XSS
  for (const pattern of XSS_PATTERNS) {
    if (pattern.test(value)) {
      return { valid: false, error: "Environment variable value contains invalid characters." }
    }
  }
  // Cek SQL injection
  for (const pattern of SQL_PATTERNS) {
    if (pattern.test(value)) {
      return { valid: false, error: "Environment variable value contains invalid characters." }
    }
  }
  return { valid: true, error: "" }
}

/**
 * Validasi satu baris env var (KEY=value).
 */
export function validateEnvLine(line: string): ValidationResult {
  const trimmed = line.trim()
  if (!trimmed) {
    return { valid: true, error: "" } // empty lines are ok
  }
  const eqIdx = trimmed.indexOf("=")
  if (eqIdx <= 0) {
    return { valid: false, error: `Invalid format: "${trimmed}" — use KEY=value format.` }
  }
  const key = trimmed.slice(0, eqIdx).trim()
  const value = trimmed.slice(eqIdx + 1).trim()
  const keyResult = validateEnvKey(key)
  if (!keyResult.valid) return keyResult
  const valueResult = validateEnvValue(value)
  if (!valueResult.valid) return valueResult
  return { valid: true, error: "" }
}

export default {
  validateProjectName,
  validateEnvKey,
  validateEnvValue,
  validateEnvLine,
  sanitizeInput,
}