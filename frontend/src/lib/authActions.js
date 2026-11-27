// Handler logout bersama untuk Nav di semua halaman (termasuk public/info).

import { getRoute, navigate } from './router.svelte.js';
import { signOut } from './session.svelte.js';
import { resetWorkspace } from './workspace.svelte.js';

export async function logout({ goHome = true, all = false } = {}) {
  // signOut already goes through invalidateSession(), which clears session
  // state and calls resetWorkspace(). Keep the explicit reset for callers
  // that need to be independent of the session handler registration.
  await signOut({ all });
  resetWorkspace();
  if (goHome && getRoute() !== '/') navigate('/');
}
