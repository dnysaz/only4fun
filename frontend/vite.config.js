import { defineConfig, loadEnv } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '');

  // Dev/preview mem-proxy /api dan /auth ke backend Go sehingga cookie sesi
  // tetap same-origin (tanpa perlu ubah CORS backend).
  const apiTarget = env.VITE_API_PROXY || 'http://localhost:8080';

  const proxy = {
    '/api': { target: apiTarget, changeOrigin: false },
    '/auth': { target: apiTarget, changeOrigin: false },
  };

  return {
    plugins: [svelte()],
    server: { proxy },
    preview: { proxy },
  };
});
