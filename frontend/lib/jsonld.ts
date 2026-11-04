/**
 * Serialisasi JSON-LD untuk dimasukkan lewat dangerouslySetInnerHTML.
 * `<` di-escape supaya nilai apa pun tidak bisa menutup tag <script> lebih awal.
 */
export function toJsonLd(data: unknown): string {
  return JSON.stringify(data).replace(/</g, "\\u003c")
}
