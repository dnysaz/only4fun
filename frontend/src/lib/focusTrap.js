// Small focus-loop helper for dialogs (no native <dialog> dependency).

const FOCUSABLE =
  'a[href], button:not([disabled]), input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])';

export function focusableElements(container) {
  if (!container) return [];
  return Array.from(container.querySelectorAll(FOCUSABLE)).filter(
    (element) => element.getClientRects().length > 0 && !element.hasAttribute('hidden'),
  );
}

// Returns true when Tab was handled inside the container.
export function trapTab(event, container) {
  if (!container || event.key !== 'Tab') return false;

  const elements = focusableElements(container);
  if (!elements.length) {
    event.preventDefault();
    return true;
  }

  const first = elements[0];
  const last = elements[elements.length - 1];
  const active = document.activeElement;

  if (!container.contains(active)) {
    event.preventDefault();
    (event.shiftKey ? last : first).focus();
    return true;
  }
  if (event.shiftKey && active === first) {
    event.preventDefault();
    last.focus();
    return true;
  }
  if (!event.shiftKey && active === last) {
    event.preventDefault();
    first.focus();
    return true;
  }
  return false;
}
