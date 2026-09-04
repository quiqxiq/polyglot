/**
 * Polyglot Web Dial Layout Engine
 * Merender navigasi Sidebar, Header, dan Router Selector secara dinamis
 * Menerapkan perizinan menu: Menu 'Pengaturan' HANYA muncul untuk role Admin.
 */

const Layout = {
  menus: [
    {
      title: 'Dashboard',
      href: 'index.html',
      icon: `<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-house"><path d="M15 21v-8a1 1 0 0 0-1-1h-4a1 1 0 0 0-1 1v8"/><path d="M3 10a2 2 0 0 1 .709-1.528l7-6a2 2 0 0 1 2.582 0l7 6A2 2 0 0 1 21 10v9a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/></svg>`,
      adminOnly: false
    },
    {
      title: 'PPP Secrets',
      href: 'ppp-secrets.html',
      icon: `<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-key-round"><path d="M2 18v3c0 .6.4 1 1 1h4v-3h3v-3h2l1.4-1.4a6.5 6.5 0 1 0-4-4Z"/><circle cx="16.5" cy="7.5" r=".5" fill="currentColor"/></svg>`,
      adminOnly: false
    },
    {
      title: 'PPP Aktif',
      href: 'ppp-active.html',
      icon: `<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-wifi"><path d="M12 20h.01"/><path d="M2 8.82a15 15 0 0 1 20 0"/><path d="M5 12.859a10 10 0 0 1 14 0"/><path d="M8.5 16.429a5 5 0 0 1 7 0"/></svg>`,
      adminOnly: false
    },
    {
      title: 'PPP Non-Aktif',
      href: 'ppp-non-active.html',
      icon: `<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-wifi-off"><path d="M12 20h.01"/><path d="M8.5 16.429a5 5 0 0 1 7 0"/><path d="M5 12.859a10 10 0 0 1 5.17-2.69"/><path d="M19 12.859a10 10 0 0 0-2.007-1.523"/><path d="M2 8.82a15 15 0 0 1 4.177-2.643"/><path d="M22 8.82a15 15 0 0 0-11.288-3.764"/><path d="m2 2 20 20"/></svg>`,
      adminOnly: false
    },
    {
      title: 'PPP Profiles',
      href: 'ppp-profiles.html',
      icon: `<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-layers"><path d="m12.83 2.18a2 2 0 0 0-1.66 0L2.6 6.08a1 1 0 0 0 0 1.83l8.58 3.91a2 2 0 0 0 1.66 0l8.58-3.9a1 1 0 0 0 0-1.83Z"/><path d="m22 17.65-9.17 4.16a2 2 0 0 1-1.66 0L2 17.65"/><path d="m22 12.65-9.17 4.16a2 2 0 0 1-1.66 0L2 12.65"/></svg>`,
      adminOnly: false
    },
    {
      title: 'Log MikroTik',
      href: 'mikrotik-logs.html',
      icon: `<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-scroll-text"><path d="M15 12h-5"/><path d="M15 8h-5"/><path d="M19 17V5a2 2 0 0 0-2-2H4"/><path d="M8 21h12a2 2 0 0 0 2-2v-1a1 1 0 0 0-1-1H11a1 1 0 0 0-1 1v2a1 1 0 0 0 1 1z"/></svg>`,
      adminOnly: false
    },
    {
      title: 'Monitor Interface',
      href: 'monitor-interface.html',
      icon: `<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-network"><rect x="16" y="16" width="6" height="6" rx="1"/><rect x="2" y="16" width="6" height="6" rx="1"/><rect x="9" y="2" width="6" height="6" rx="1"/><path d="M5 16v-3a1 1 0 0 1 1-1h12a1 1 0 0 1 1 1v3"/><path d="M12 12V8"/></svg>`,
      adminOnly: false
    },
    {
      title: 'Pengaturan',
      href: 'settings.html',
      icon: `<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-settings"><path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z"/><circle cx="12" cy="12" r="3"/></svg>`,
      adminOnly: true // KHUSUS ADMIN
    }
  ],

  pingStreamController: null,
  currentPingDeviceId: null,

  init(activePageHref = 'index.html') {
    const user = Auth.getUser() || { username: 'Pengguna', role: 'teknisi' };
    const isAdmin = Auth.isAdmin();

    this.renderSidebar(activePageHref, isAdmin);
    this.renderHeader(user);
    this.renderMobileNav(activePageHref, user, isAdmin);
    this.bindEvents();

    // Inisialisasi daftar router dan streaming ping global di header
    RouterSelector.init();

    // Segera mulai stream ping jika ID router sudah ada di memory / localStorage
    const initialDeviceId = RouterSelector.getSelectedDeviceId();
    if (initialDeviceId) {
      this.startHeaderPingStream(initialDeviceId);
    }

    RouterSelector.onDeviceChange((deviceId) => {
      if (!deviceId) return;
      this.startHeaderPingStream(deviceId);
    });

    window.addEventListener('beforeunload', () => {
      this.cleanupPingStream();
    });
  },

  cleanupPingStream() {
    if (this.pingStreamController) {
      this.pingStreamController.abort();
      this.pingStreamController = null;
    }
    this.currentPingDeviceId = null;
  },

  startHeaderPingStream(deviceId) {
    if (!deviceId) return;
    if (this.currentPingDeviceId === deviceId && this.pingStreamController) {
      return;
    }
    this.cleanupPingStream();
    this.currentPingDeviceId = deviceId;

    const setUI = (text, status) => {
      const valEl = document.getElementById('header-ping-val');
      const dotEl = document.getElementById('header-ping-dot');
      const pingEl = document.getElementById('header-ping-ping');
      const mobVal = document.getElementById('mobile-header-ping-val');
      const mobDot = document.getElementById('mobile-header-ping-dot');

      if (valEl) valEl.textContent = text;
      if (mobVal) mobVal.textContent = text;

      let dotColor = 'bg-emerald-500';
      let pingColor = 'bg-emerald-400';
      let textColor = 'text-emerald-700';

      if (status === 'timeout' || status === 'error') {
        dotColor = 'bg-rose-500';
        pingColor = 'bg-rose-400';
        textColor = 'text-rose-600';
      } else if (status === 'high') {
        dotColor = 'bg-rose-500';
        pingColor = 'bg-rose-400';
        textColor = 'text-rose-600';
      } else if (status === 'medium') {
        dotColor = 'bg-amber-500';
        pingColor = 'bg-amber-400';
        textColor = 'text-amber-600';
      }

      if (valEl) valEl.className = `font-bold ${textColor}`;
      if (mobVal) mobVal.className = `font-bold ${textColor}`;
      if (dotEl) dotEl.className = `relative inline-flex rounded-full h-2.5 w-2.5 ${dotColor}`;
      if (mobDot) mobDot.className = `relative inline-flex rounded-full h-2 w-2 ${dotColor}`;
      if (pingEl) pingEl.className = `animate-ping absolute inline-flex h-full w-full rounded-full ${pingColor} opacity-75`;
    };

    this.pingStreamController = new AbortController();
    const signal = this.pingStreamController.signal;

    setUI('-- ms', 'medium');

    API.stream(
      '/polyglot.v1.DeviceService/StreamPing',
      { id: deviceId, address: '8.8.8.8' },
      (frame) => {
        const latency = Number(frame.latencyMs !== undefined ? frame.latencyMs : 0);
        const isAlive = frame.status !== 'timeout' &&
                        frame.status !== 'host unreachable' &&
                        frame.status !== 'net unreachable';

        if (!isAlive) {
          setUI('Timeout', 'timeout');
        } else {
          const status = latency > 100 ? 'high' : latency > 50 ? 'medium' : 'good';
          setUI(`${latency} ms`, status);
        }
      },
      (err) => {
        if (signal.aborted) return;
        setUI('-- ms', 'error');
        setTimeout(() => {
          if (!signal.aborted) this.startHeaderPingStream(deviceId);
        }, 5000);
      },
      signal
    );
  },

  renderSidebar(activePage, isAdmin) {
    const container = document.getElementById('sidebar-container');
    if (!container) return;

    const navLinks = this.menus
      .filter(m => !m.adminOnly || isAdmin)
      .map(m => {
        const isActive = activePage === m.href || (activePage === '' && m.href === 'index.html');
        return `
          <a href="${m.href}" class="flex items-center px-6 py-3.5 gap-3 font-medium transition-all duration-200 ${
            isActive
              ? 'text-blue-600 bg-blue-50/80 border-r-[3px] border-blue-600 shadow-sm'
              : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900'
          }">
            ${m.icon}
            <span class="text-sm">${m.title}</span>
            ${m.adminOnly ? '<span class="ml-auto text-[10px] uppercase font-bold tracking-wider px-1.5 py-0.5 rounded bg-purple-100 text-purple-700">Admin</span>' : ''}
          </a>
        `;
      }).join('');

    container.innerHTML = `
      <div class="hidden lg:block fixed top-0 left-0 h-screen w-64 bg-white border-r border-gray-200 z-40 overflow-y-auto flex flex-col shadow-sm">
        <!-- Logo Header -->
        <div class="flex items-center justify-between h-16 px-6 bg-gradient-to-r from-blue-600 to-indigo-600 text-white shadow">
          <div class="flex items-center gap-2.5">
            <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-router"><rect width="20" height="8" x="2" y="14" rx="2"/><path d="M6.01 18h.01"/><path d="M10.01 18h.01"/><path d="M15 10v4"/><path d="M17.84 7.17a4 4 0 0 0-5.66 0"/><path d="M20.66 4.34a8 8 0 0 0-11.31 0"/></svg>
            <span class="font-bold text-lg tracking-tight">GNet Dial</span>
          </div>
          <span class="text-[10px] font-semibold bg-white/20 px-2 py-0.5 rounded-full uppercase tracking-wider">v2.0</span>
        </div>

        <!-- Nav Items -->
        <nav class="mt-4 flex-1 space-y-1">
          ${navLinks}
        </nav>

        <!-- Logout Section -->
        <div class="p-4 border-t border-gray-100">
          <button id="sidebar-logout-btn" class="flex items-center w-full px-4 py-2.5 text-sm font-medium text-red-600 hover:bg-red-50 rounded-lg transition-colors gap-2.5">
            <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-log-out"><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/><polyline points="16 17 21 12 16 7"/><line x1="21" x2="9" y1="12" y2="12"/></svg>
            <span>Keluar Portal</span>
          </button>
        </div>
      </div>
    `;
  },

  renderHeader(user) {
    const container = document.getElementById('header-container');
    if (!container) return;

    const roleName = (user.roles && user.roles.length > 0) ? user.roles[0] : (user.role || 'Teknisi');
    const isAdmin = Auth.isAdmin();

    container.innerHTML = `
      <header class="hidden lg:flex fixed top-0 right-0 left-64 h-16 bg-white/95 backdrop-blur-sm border-b border-gray-200 px-6 z-30 items-center justify-between shadow-xs">
        <!-- Router Selector & Ping Badge -->
        <div class="flex items-center gap-4">
          <div class="flex items-center gap-2 text-sm text-gray-500 font-medium">
            <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-blue-600"><rect width="20" height="8" x="2" y="14" rx="2"/><path d="M6.01 18h.01"/><path d="M10.01 18h.01"/><path d="M15 10v4"/></svg>
            <span>Router:</span>
          </div>
          <select class="router-select-control bg-gray-50 border border-gray-300 text-gray-800 text-sm rounded-lg focus:ring-blue-500 focus:border-blue-500 block px-3 py-1.5 font-medium outline-none transition-shadow cursor-pointer min-w-[220px]">
            <option value="">Memuat router...</option>
          </select>

          <!-- Streaming Ping Badge di Header (Minimalist ms only) -->
          <div id="header-ping-container" class="flex items-center gap-2 px-3 py-1.5 rounded-xl bg-slate-100/90 border border-slate-200 text-xs font-mono shadow-xs" title="Latensi Ping MikroTik ke Internet (8.8.8.8)">
            <span class="flex h-2.5 w-2.5 relative">
              <span id="header-ping-ping" class="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
              <span id="header-ping-dot" class="relative inline-flex rounded-full h-2.5 w-2.5 bg-emerald-500"></span>
            </span>
            <span id="header-ping-val" class="font-bold text-emerald-700 tracking-tight">-- ms</span>
          </div>
        </div>

        <!-- User Identity -->
        <div class="flex items-center gap-4">
          <div class="flex items-center gap-2.5">
            <div class="w-8 h-8 rounded-full bg-gradient-to-tr from-blue-600 to-indigo-600 text-white flex items-center justify-center font-bold text-xs uppercase shadow-xs">
              ${(user.username || 'U')[0]}
            </div>
            <div class="flex flex-col text-right">
              <span class="text-sm font-semibold text-gray-800 leading-tight">${user.full_name || user.username}</span>
              <span class="text-[11px] font-medium uppercase tracking-wider ${isAdmin ? 'text-purple-600' : 'text-blue-600'}">${roleName}</span>
            </div>
          </div>
        </div>
      </header>
    `;
  },

  renderMobileNav(activePage, user, isAdmin) {
    const container = document.getElementById('mobile-nav-container');
    if (!container) return;

    const navLinks = this.menus
      .filter(m => !m.adminOnly || isAdmin)
      .map(m => {
        const isActive = activePage === m.href || (activePage === '' && m.href === 'index.html');
        return `
          <a href="${m.href}" class="flex items-center px-4 py-3 gap-3 rounded-lg font-medium text-sm transition-colors ${
            isActive ? 'text-blue-600 bg-blue-50 font-semibold' : 'text-gray-700 hover:bg-gray-100'
          }">
            ${m.icon}
            <span>${m.title}</span>
            ${m.adminOnly ? '<span class="ml-auto text-[9px] uppercase font-bold tracking-wider px-1.5 py-0.5 rounded bg-purple-100 text-purple-700">Admin</span>' : ''}
          </a>
        `;
      }).join('');

    container.innerHTML = `
      <!-- Mobile Top Bar -->
      <div class="lg:hidden fixed top-0 left-0 right-0 h-16 bg-white border-b border-gray-200 px-4 z-40 flex items-center justify-between shadow-xs">
        <div class="flex items-center gap-2">
          <button id="mobile-menu-open-btn" class="p-2 text-gray-600 hover:text-gray-900 rounded-lg hover:bg-gray-100">
            <svg xmlns="http://www.w3.org/2000/svg" width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="4" x2="20" y1="12" y2="12"/><line x1="4" x2="20" y1="6" y2="6"/><line x1="4" x2="20" y1="18" y2="18"/></svg>
          </button>
          <span class="font-bold text-base bg-gradient-to-r from-blue-600 to-indigo-600 bg-clip-text text-transparent">GNet Dial</span>
        </div>

        <!-- Mobile Ping Badge -->
        <div class="flex items-center gap-1.5 px-2.5 py-1 rounded-lg bg-slate-100 border border-slate-200 text-xs font-mono shadow-xs" title="Latensi Ping MikroTik ke Internet (8.8.8.8)">
          <span class="flex h-2 w-2 relative">
            <span id="mobile-header-ping-dot" class="relative inline-flex rounded-full h-2 w-2 bg-emerald-500"></span>
          </span>
          <span id="mobile-header-ping-val" class="font-bold text-emerald-700">-- ms</span>
        </div>

        <button id="mobile-logout-btn" class="p-2 text-red-500 hover:bg-red-50 rounded-lg">
          <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/><polyline points="16 17 21 12 16 7"/><line x1="21" x2="9" y1="12" y2="12"/></svg>
        </button>
      </div>

      <!-- Mobile Sidebar Drawer -->
      <div id="mobile-drawer" class="lg:hidden fixed inset-0 z-50 hidden">
        <div id="mobile-drawer-backdrop" class="fixed inset-0 bg-slate-900/60 backdrop-blur-xs transition-opacity"></div>
        <div class="fixed top-0 bottom-0 left-0 w-72 bg-white flex flex-col shadow-2xl z-10 transform -translate-x-full transition-transform duration-300" id="mobile-drawer-panel">
          <div class="flex items-center justify-between h-16 px-4 bg-gradient-to-r from-blue-600 to-indigo-600 text-white shadow">
            <span class="font-bold text-base">GNet Dial Menu</span>
            <button id="mobile-menu-close-btn" class="p-1.5 hover:bg-white/20 rounded-md">
              <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 6 6 18"/><path d="m6 6 12 12"/></svg>
            </button>
          </div>

          <!-- Router Selector in Mobile -->
          <div class="p-4 border-b border-gray-100 bg-gray-50">
            <label class="block text-xs font-semibold text-gray-500 uppercase tracking-wider mb-1.5">Pilih Router</label>
            <select class="router-select-control w-full bg-white border border-gray-300 text-gray-800 text-xs rounded-lg px-2.5 py-2 font-medium outline-none">
              <option value="">Memuat router...</option>
            </select>
          </div>

          <!-- Links -->
          <nav class="p-3 flex-1 overflow-y-auto space-y-1">
            ${navLinks}
          </nav>

          <!-- User Footer -->
          <div class="p-4 border-t border-gray-100 bg-gray-50 flex items-center justify-between">
            <div class="flex items-center gap-2">
              <div class="w-8 h-8 rounded-full bg-blue-600 text-white flex items-center justify-center font-bold text-xs uppercase">
                ${(user.username || 'U')[0]}
              </div>
              <div class="flex flex-col">
                <span class="text-xs font-semibold text-gray-800">${user.username}</span>
                <span class="text-[10px] text-gray-500 capitalize">${user.role || 'Teknisi'}</span>
              </div>
            </div>
            <button class="text-xs text-red-600 font-semibold hover:underline" onclick="Auth.logout()">Keluar</button>
          </div>
        </div>
      </div>
    `;
  },

  bindEvents() {
    // Logout buttons
    const sideLogout = document.getElementById('sidebar-logout-btn');
    if (sideLogout) sideLogout.onclick = () => Auth.logout();

    const mobLogout = document.getElementById('mobile-logout-btn');
    if (mobLogout) mobLogout.onclick = () => Auth.logout();

    // Mobile Drawer toggling
    const openBtn = document.getElementById('mobile-menu-open-btn');
    const closeBtn = document.getElementById('mobile-menu-close-btn');
    const drawer = document.getElementById('mobile-drawer');
    const backdrop = document.getElementById('mobile-drawer-backdrop');
    const panel = document.getElementById('mobile-drawer-panel');

    const openDrawer = () => {
      if (!drawer) return;
      drawer.classList.remove('hidden');
      requestAnimationFrame(() => {
        panel.classList.remove('-translate-x-full');
      });
    };

    const closeDrawer = () => {
      if (!drawer) return;
      panel.classList.add('-translate-x-full');
      setTimeout(() => drawer.classList.add('hidden'), 300);
    };

    if (openBtn) openBtn.onclick = openDrawer;
    if (closeBtn) closeBtn.onclick = closeDrawer;
    if (backdrop) backdrop.onclick = closeDrawer;
  }
};

window.Layout = Layout;
