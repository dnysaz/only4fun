// Status visibility halaman dipakai bersama supaya polling berhenti saat tab
// tidak terlihat. Satu listener untuk seluruh aplikasi, bukan per komponen.

export const visibility = $state({
  visible: typeof document === 'undefined' ? true : !document.hidden,
});

let started = false;

export function watchPageVisibility() {
  if (started) return () => {};
  started = true;

  function sync() {
    visibility.visible = !document.hidden;
  }

  document.addEventListener('visibilitychange', sync);
  window.addEventListener('pageshow', sync);
  sync();

  return () => {
    document.removeEventListener('visibilitychange', sync);
    window.removeEventListener('pageshow', sync);
    started = false;
  };
}
