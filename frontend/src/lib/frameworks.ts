// Data ikon resmi dari simple-icons. Path-nya adalah logo asli tiap framework,
// bukan gambar buatan sendiri, jadi bentuknya tetap akurat di ukuran besar.
import {
  siAngular,
  siAstro,
  siHtml5,
  siJavascript,
  siNextdotjs,
  siNuxt,
  siReact,
  siSolid,
  siSvelte,
  siVite,
  siVuedotjs,
} from 'simple-icons';

export interface Framework {
  id: string;
  label: string;
  path: string;
  color: string;
}

export const frameworks: Framework[] = [
  { id: 'next', label: 'Next.js', path: siNextdotjs.path, color: `#${siNextdotjs.hex}` },
  { id: 'astro', label: 'Astro', path: siAstro.path, color: `#${siAstro.hex}` },
  { id: 'vue', label: 'Vue', path: siVuedotjs.path, color: `#${siVuedotjs.hex}` },
  { id: 'react', label: 'React', path: siReact.path, color: `#${siReact.hex}` },
  { id: 'svelte', label: 'Svelte', path: siSvelte.path, color: `#${siSvelte.hex}` },
  { id: 'html', label: 'Static HTML', path: siHtml5.path, color: `#${siHtml5.hex}` },
  { id: 'vite', label: 'Vite', path: siVite.path, color: `#${siVite.hex}` },
  { id: 'nuxt', label: 'Nuxt', path: siNuxt.path, color: `#${siNuxt.hex}` },
  // simple-icons menyimpan Angular sebagai monokrom; warna mereknya tetap merah.
  { id: 'angular', label: 'Angular', path: siAngular.path, color: '#dd0031' },
  { id: 'solid', label: 'Solid', path: siSolid.path, color: `#${siSolid.hex}` },
  { id: 'js', label: 'and any other JS framework', path: siJavascript.path, color: `#${siJavascript.hex}` },
];
