/**
 * Polyglot Web Dial Authentication & Capability System
 *
 * Login admission and page access are driven by the effective permission list
 * returned by the backend (the same list Casbin uses). There is no hardcoded
 * role allowlist: a role gets Dial access by being granted permissions that map
 * to at least one visible menu.
 */

const Auth = {
  TOKEN_KEY: 'polyglot_token',
  USER_KEY: 'polyglot_user',
  EXPIRES_KEY: 'polyglot_token_expires_at',
  SYNCED_KEY: 'polyglot_user_synced_at',

  /**
   * Authenticates against the shared AuthService and admits the user only when
   * their permissions unlock at least one Dial menu.
   */
  async login(username, password) {
    const res = await API.call('/polyglot.v1.AuthService/Login', {
      username: username.trim(),
      password: password
    }, false);

    const user = res.user || {};
    const permissions = this.getPermissions(user);

    if (typeof DialMenu === 'undefined') {
      throw new Error('Menu registry tidak termuat. Muat ulang halaman.');
    }

    if (!DialMenu.hasAnyMenu((required) => DialMenu.hasPermission(permissions, required))) {
      throw new Error('Akses Ditolak: Akun Anda belum memiliki izin untuk membuka Portal Dial. Hubungi Admin untuk mendapatkan akses.');
    }

    localStorage.setItem(this.TOKEN_KEY, res.token);
    localStorage.setItem(this.USER_KEY, JSON.stringify(user));
    localStorage.setItem(this.EXPIRES_KEY, String(res.expiresAtUnix || 0));
    localStorage.setItem(this.SYNCED_KEY, String(Date.now()));
    return user;
  },

  logout() {
    try {
      API.call('/polyglot.v1.AuthService/Logout', {}, true).catch(() => {});
    } catch (_) {}

    this.clearSession();
    window.location.href = 'login.html';
  },

  clearSession() {
    localStorage.removeItem(this.TOKEN_KEY);
    localStorage.removeItem(this.USER_KEY);
    localStorage.removeItem(this.EXPIRES_KEY);
    localStorage.removeItem(this.SYNCED_KEY);
    localStorage.removeItem('selected_device_id');
  },

  getUser() {
    try {
      const u = localStorage.getItem(this.USER_KEY);
      return u ? JSON.parse(u) : null;
    } catch {
      return null;
    }
  },

  getPermissions(user) {
    const source = user || this.getUser();
    if (!source || !Array.isArray(source.permissions)) return [];
    return source.permissions;
  },

  /**
   * Replicates the backend Casbin matcher against the flattened permission list
   * returned by GetImplicitPermissionsForUser ("resource:action", action "*").
   */
  can(required) {
    if (typeof DialMenu === 'undefined') return false;
    return DialMenu.hasPermission(this.getPermissions(), required);
  },

  canAll(required) {
    if (typeof DialMenu === 'undefined') return false;
    return DialMenu.canAll(this.getPermissions(), required);
  },

  canAny(required) {
    if (typeof DialMenu === 'undefined') return false;
    return DialMenu.canAny(this.getPermissions(), required);
  },

  permissionChecker() {
    return (required) => this.can(required);
  },

  canMenu(menuId) {
    if (typeof DialMenu === 'undefined') return false;
    return DialMenu.canAccess(menuId, this.permissionChecker());
  },

  canAction(menuId, actionKey) {
    if (typeof DialMenu === 'undefined') return false;
    return DialMenu.canAction(menuId, actionKey, this.permissionChecker());
  },

  /**
   * Tries to map the current account to a menu id. Accepts a menu id or a page
   * href (e.g. "ppp-active.html").
   */
  resolveMenu(menuIdOrHref) {
    if (typeof DialMenu === 'undefined') return null;
    return DialMenu.findNode(menuIdOrHref) || DialMenu.findPageByHref(menuIdOrHref);
  },

  async refreshUser() {
    const res = await API.call('/polyglot.v1.AuthService/GetMe', {});
    if (res && res.user) {
      localStorage.setItem(this.USER_KEY, JSON.stringify(res.user));
      localStorage.setItem(this.SYNCED_KEY, String(Date.now()));
      return res.user;
    }
    return this.getUser();
  },

  /**
   * Refreshes effective permissions from the server so revocations apply
   * without waiting for the next login. Falls back to the cached profile when
   * the account cannot call GetMe (no profile:read) or the network is offline.
   */
  async ensureFreshUser(maxAgeMs = 60_000) {
    if (!this.hasSession()) return null;
    if (!this.can('profile:read')) return this.getUser();

    const lastSync = Number(localStorage.getItem(this.SYNCED_KEY) || 0);
    if (lastSync > 0 && Date.now() - lastSync < maxAgeMs) {
      return this.getUser();
    }

    try {
      return await this.refreshUser();
    } catch (err) {
      if (err && (err.status === 401 || err.permissionDenied)) {
        return this.getUser();
      }
      console.warn('Failed to refresh user permissions:', err);
      return this.getUser();
    }
  },

  hasSession() {
    const token = localStorage.getItem(this.TOKEN_KEY);
    const user = this.getUser();
    if (!token || !user) return false;

    const expiresAt = Number(localStorage.getItem(this.EXPIRES_KEY) || 0);
    if (expiresAt > 0 && Date.now() >= expiresAt * 1000) {
      this.clearSession();
      return false;
    }
    return true;
  },

  /**
   * Synchronous guard that only verifies the session and that at least one menu
   * is available. Page-specific guards should use requireMenu().
   */
  requireAuth() {
    if (!this.hasSession()) {
      API.redirectToLogin('Silakan login terlebih dahulu');
      return false;
    }

    if (typeof DialMenu !== 'undefined' && !DialMenu.hasAnyMenu(this.permissionChecker())) {
      this.redirectNoAccess();
      return false;
    }
    return true;
  },

  /**
   * Async page guard: refreshes permissions, validates every ancestor of the
   * requested menu, and redirects to the first accessible page otherwise.
   */
  async requireMenu(menuIdOrHref) {
    if (!this.hasSession()) {
      API.redirectToLogin('Silakan login terlebih dahulu');
      return false;
    }

    await this.ensureFreshUser();

    if (!this.hasSession()) {
      API.redirectToLogin('Sesi Anda telah berakhir, silakan login kembali');
      return false;
    }

    if (typeof DialMenu === 'undefined') {
      console.error('DialMenu registry is not loaded');
      return false;
    }

    const can = this.permissionChecker();

    if (!DialMenu.hasAnyMenu(can)) {
      this.redirectNoAccess();
      return false;
    }

    const node = this.resolveMenu(menuIdOrHref);
    if (!node || !DialMenu.canAccess(node.id, can)) {
      const target = DialMenu.firstAccessibleRoute(can);
      if (target && target !== 'no-access.html') {
        window.location.replace(target);
      } else {
        this.redirectNoAccess();
      }
      return false;
    }

    return true;
  },

  redirectNoAccess() {
    if (window.location.pathname.endsWith('no-access.html')) return;
    window.location.replace('no-access.html');
  },

  firstAccessibleRoute() {
    if (typeof DialMenu === 'undefined') return null;
    return DialMenu.firstAccessibleRoute(this.permissionChecker());
  },

  isAdmin() {
    const u = this.getUser();
    if (!u) return false;
    const roles = u.roles && u.roles.length > 0 ? u.roles : [u.role];
    return roles.some((r) => r && r.toLowerCase() === 'admin');
  },

  primaryRole(user) {
    const u = user || this.getUser() || {};
    return (u.roles && u.roles.length > 0) ? u.roles[0] : (u.role || 'user');
  }
};

/**
 * Toast Notification Utility
 */
function showToast(message, type = 'info') {
  let container = document.getElementById('toast-container');
  if (!container) {
    container = document.createElement('div');
    container.id = 'toast-container';
    container.className = 'fixed top-4 right-4 z-50 flex flex-col gap-2 pointer-events-none';
    document.body.appendChild(container);
  }

  const toast = document.createElement('div');
  toast.className = `pointer-events-auto flex items-center gap-3 px-4 py-3 rounded-xl shadow-lg text-sm font-medium transition-all duration-300 transform translate-y-[-10px] opacity-0 ${
    type === 'success' ? 'bg-emerald-600 text-white' :
    type === 'error' ? 'bg-rose-600 text-white' :
    type === 'warning' ? 'bg-amber-500 text-white' :
    'bg-slate-800 text-white'
  }`;

  toast.innerHTML = `
    <span>${message}</span>
    <button class="ml-auto opacity-75 hover:opacity-100 text-base leading-none">&times;</button>
  `;

  const closeBtn = toast.querySelector('button');
  closeBtn.onclick = () => {
    toast.classList.add('opacity-0', 'translate-y-[-10px]');
    setTimeout(() => toast.remove(), 300);
  };

  container.appendChild(toast);
  requestAnimationFrame(() => {
    toast.classList.remove('opacity-0', 'translate-y-[-10px]');
  });

  setTimeout(() => {
    if (toast.parentElement) {
      toast.classList.add('opacity-0', 'translate-y-[-10px]');
      setTimeout(() => toast.remove(), 300);
    }
  }, 3500);
}

window.Auth = Auth;
window.showToast = showToast;
