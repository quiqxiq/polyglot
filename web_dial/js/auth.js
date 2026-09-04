/**
 * Polyglot Web Dial Authentication & Guard System
 * Menangani login, session token, dan pembatasan role (Admin & Teknisi saja)
 */

const Auth = {
  async login(username, password) {
    const res = await API.call('/polyglot.v1.AuthService/Login', {
      username: username.trim(),
      password: password
    }, false);

    const user = res.user || {};
    const effectiveRoles = (user.roles && user.roles.length > 0) ? user.roles : [user.role];
    const isOwnerOrSuper = effectiveRoles.some(r => r.toLowerCase() === 'owner' || r.toLowerCase() === 'superadmin');

    if (isOwnerOrSuper) {
      throw new Error('Akses Ditolak: Portal Dial dikhususkan untuk Admin dan Teknisi. Akun Superadmin/Owner silakan gunakan Portal Utama di web admin.');
    }

    const isAdminOrTeknisi = effectiveRoles.some(r => ['admin', 'teknisi', 'technician', 'agent'].includes(r.toLowerCase()));
    if (!isAdminOrTeknisi) {
      throw new Error('Akses Ditolak: Role akun Anda tidak diizinkan mengakses Portal Dial.');
    }

    localStorage.setItem('polyglot_token', res.token);
    localStorage.setItem('polyglot_user', JSON.stringify(user));
    return user;
  },

  logout() {
    try {
      API.call('/polyglot.v1.AuthService/Logout', {}, true).catch(() => {});
    } catch (_) {}

    localStorage.removeItem('polyglot_token');
    localStorage.removeItem('polyglot_user');
    localStorage.removeItem('selected_device_id');
    window.location.href = 'login.html';
  },

  getUser() {
    try {
      const u = localStorage.getItem('polyglot_user');
      return u ? JSON.parse(u) : null;
    } catch {
      return null;
    }
  },

  isAdmin() {
    const u = this.getUser();
    if (!u) return false;
    const roles = u.roles && u.roles.length > 0 ? u.roles : [u.role];
    return roles.some(r => r && r.toLowerCase() === 'admin');
  },

  requireAuth(allowedRoles = ['admin', 'teknisi', 'technician', 'agent']) {
    const token = localStorage.getItem('polyglot_token');
    const user = this.getUser();

    if (!token || !user) {
      API.redirectToLogin('Silakan login terlebih dahulu');
      return false;
    }

    const effectiveRoles = (user.roles && user.roles.length > 0) ? user.roles : [user.role];
    const hasRole = effectiveRoles.some(r => r && allowedRoles.map(x => x.toLowerCase()).includes(r.toLowerCase()));

    if (!hasRole) {
      alert('Akses Ditolak: Anda tidak memiliki izin untuk membuka halaman ini.');
      window.location.href = 'index.html';
      return false;
    }

    return true;
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
