<script>
  // Renderer generik: semua bentuk/icon berasal dari path resmi simple-icons
  // di frameworks.ts. Tidak ada lagi logo gambar sendiri yang bentuknya kacau.
  let { framework, size = 76 } = $props();
</script>

<svg
  viewBox="0 0 24 24"
  width={size}
  height={size}
  fill={framework.color}
  role="img"
  aria-label={framework.label}
  xmlns="http://www.w3.org/2000/svg"
>
  <path d={framework.path} />
</svg>

<style>
  svg {
    display: block;
  }
</style>
