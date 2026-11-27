// Copy helper with a legacy fallback. Returns true only when the browser
// confirms the copy succeeded, so callers can surface a visible failure.

export async function copyText(value) {
  if (!value) return false;

  if (navigator.clipboard?.writeText) {
    try {
      await navigator.clipboard.writeText(value);
      return true;
    } catch {
      // Fall through to the textarea fallback.
    }
  }

  const helper = document.createElement('textarea');
  helper.value = String(value);
  helper.setAttribute('readonly', '');
  helper.style.position = 'fixed';
  helper.style.top = '0';
  helper.style.left = '0';
  helper.style.opacity = '0';
  document.body.appendChild(helper);
  helper.select();

  let copied = false;
  try {
    copied = document.execCommand('copy');
  } catch {
    copied = false;
  } finally {
    helper.remove();
  }
  return copied === true;
}
