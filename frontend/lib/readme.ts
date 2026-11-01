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
