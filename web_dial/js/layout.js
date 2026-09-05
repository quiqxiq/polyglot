/**
 * Polyglot Web Dial Layout Engine
 * Merender navigasi Sidebar, Centered Header, Mobile Top Bar, dan Bottom Navigation Bar secara dinamis.
 * Menerapkan perizinan menu: Menu 'Pengaturan' HANYA muncul untuk role Admin.
 */

const Layout = {
  menus: [
    {
      title: 'Dashboard',
      href: 'index.html',
      icon: `<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-house"><path d="M15 21v-8a1 1 0 0 0-1-1h-4a1 1 0 0 0-1 1v8"/><path d="M3 10a2 2 0 0 1 .709-1.528l7-6a2 2 0 0 1 2.582 0l7 6A2 2 0 0 1 21 10v9a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/></svg>`,
      group: 'operasional',
      bottomNav: true,
      adminOnly: false
    },
    {
      title: 'PPP Aktif',
      href: 'ppp-active.html',
      icon: `<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-wifi"><path d="M12 20h.01"/><path d="M2 8.82a15 15 0 0 1 20 0"/><path d="M5 12.859a10 10 0 0 1 14 0"/><path d="M8.5 16.429a5 5 0 0 1 7 0"/></svg>`,
      group: 'operasional',
      bottomNav: true,
      adminOnly: false
    },
    {
      title: 'PPP Non-Aktif',
      href: 'ppp-non-active.html',
      icon: `<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-wifi-off"><path d="M12 20h.01"/><path d="M8.5 16.429a5 5 0 0 1 7 0"/><path d="M5 12.859a10 10 0 0 1 5.17-2.69"/><path d="M19 12.859a10 10 0 0 0-2.007-1.523"/><path d="M2 8.82a15 15 0 0 1 4.177-2.643"/><path d="M22 8.82a15 15 0 0 0-11.288-3.764"/><path d="m2 2 20 20"/></svg>`,
      group: 'operasional',
      bottomNav: true,
      adminOnly: false
    },
    {
      title: 'Log MikroTik',
      href: 'mikrotik-logs.html',
      icon: `<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-scroll-text"><path d="M15 12h-5"/><path d="M15 8h-5"/><path d="M19 17V5a2 2 0 0 0-2-2H4"/><path d="M8 21h12a2 2 0 0 0 2-2v-1a1 1 0 0 0-1-1H11a1 1 0 0 0-1 1v2a1 1 0 0 0 1 1z"/></svg>`,
      group: 'operasional',
      bottomNav: true,
      adminOnly: false
    },
    {
      title: 'Monitor Interface',
      href: 'monitor-interface.html',
      icon: `<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-network"><rect x="16" y="16" width="6" height="6" rx="1"/><rect x="2" y="16" width="6" height="6" rx="1"/><rect x="9" y="2" width="6" height="6" rx="1"/><path d="M5 16v-3a1 1 0 0 1 1-1h12a1 1 0 0 1 1 1v3"/><path d="M12 12V8"/></svg>`,
      group: 'operasional',
      bottomNav: true,
      adminOnly: false
    },
    {
      title: 'PPP Secrets',
      href: 'ppp-secrets.html',
      icon: `<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-key-round"><path d="M2 18v3c0 .6.4 1 1 1h4v-3h3v-3h2l1.4-1.4a6.5 6.5 0 1 0-4-4Z"/><circle cx="16.5" cy="7.5" r=".5" fill="currentColor"/></svg>`,
      group: 'manajemen',
      bottomNav: false,
      adminOnly: false
    },
    {
      title: 'PPP Profiles',
      href: 'ppp-profiles.html',
      icon: `<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-layers"><path d="m12.83 2.18a2 2 0 0 0-1.66 0L2.6 6.08a1 1 0 0 0 0 1.83l8.58 3.91a2 2 0 0 0 1.66 0l8.58-3.9a1 1 0 0 0 0-1.83Z"/><path d="m22 17.65-9.17 4.16a2 2 0 0 1-1.66 0L2 17.65"/><path d="m22 12.65-9.17 4.16a2 2 0 0 1-1.66 0L2 12.65"/></svg>`,
      group: 'manajemen',
      bottomNav: false,
      adminOnly: false
    },
    {
      title: 'Pengaturan',
      href: 'settings.html',
      icon: `<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-settings"><path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z"/><circle cx="12" cy="12" r="3"/></svg>`,
      group: 'sistem',
      bottomNav: false,
      adminOnly: true
    }
  ],

  pingStreamController: null,
  currentPingDeviceId: null,

  init(activePageHref = 'index.html') {
    const user = Auth.getUser() || { username: 'Pengguna', role: 'teknisi' };
    const isAdmin = Auth.isAdmin();

    this.renderSidebar(activePageHref, isAdmin, user);
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

      if (valEl) valEl.className = `font-bold font-mono ${textColor}`;
      if (mobVal) mobVal.className = `font-bold font-mono ${textColor}`;
      if (dotEl) dotEl.className = `relative inline-flex rounded-full h-2 w-2 sm:h-2.5 sm:w-2.5 ${dotColor}`;
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

  renderSidebar(activePage, isAdmin, user) {
    const container = document.getElementById('sidebar-container');
    if (!container) return;

    const renderLink = (m) => {
      const isActive = activePage === m.href || (activePage === '' && m.href === 'index.html');
      return `
        <a href="${m.href}" class="flex items-center px-3.5 py-2.5 mx-3 rounded-xl gap-3 text-sm font-medium transition-all duration-200 group ${
          isActive
            ? 'text-blue-600 bg-blue-50/90 font-semibold shadow-xs'
            : 'text-slate-600 hover:bg-slate-50 hover:text-slate-900'
        }">
          <span class="${isActive ? 'text-blue-600' : 'text-slate-400 group-hover:text-slate-600'} transition-colors">
            ${m.icon}
          </span>
          <span>${m.title}</span>
          ${m.adminOnly ? '<span class="ml-auto text-[9px] uppercase font-bold tracking-wider px-1.5 py-0.5 rounded bg-purple-50 text-purple-700 border border-purple-200/60">Admin</span>' : ''}
        </a>
      `;
    };

    const operasionalLinks = this.menus
      .filter(m => m.group === 'operasional' && (!m.adminOnly || isAdmin))
      .map(renderLink).join('');

    const manajemenLinks = this.menus
      .filter(m => m.group === 'manajemen' && (!m.adminOnly || isAdmin))
      .map(renderLink).join('');

    const sistemLinks = this.menus
      .filter(m => m.group === 'sistem' && (!m.adminOnly || isAdmin))
      .map(renderLink).join('');

    const roleName = (user.roles && user.roles.length > 0) ? user.roles[0] : (user.role || 'Teknisi');

    container.innerHTML = `
      <aside class="hidden lg:flex fixed top-0 left-0 h-screen w-64 bg-white border-r border-slate-200/80 z-40 flex-col shadow-xs select-none">
        <!-- Logo Header -->
        <div class="flex items-center justify-between h-16 px-6 border-b border-slate-100 bg-white">
          <div class="flex items-center gap-2.5">
            <div class="w-9 h-9 rounded-xl bg-gradient-to-tr from-blue-600 to-indigo-600 text-white flex items-center justify-center shadow-md shadow-blue-500/20">
              <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="20" height="8" x="2" y="14" rx="2"/><path d="M6.01 18h.01"/><path d="M10.01 18h.01"/><path d="M15 10v4"/><path d="M17.84 7.17a4 4 0 0 0-5.66 0"/><path d="M20.66 4.34a8 8 0 0 0-11.31 0"/></svg>
            </div>
            <div>
              <span class="font-bold text-base text-slate-900 tracking-tight leading-none block">GNet Dial</span>
              <span class="text-[10px] font-semibold text-slate-400 uppercase tracking-wider block mt-0.5">Operasional</span>
            </div>
          </div>
          <span class="text-[10px] font-bold bg-blue-50 text-blue-700 border border-blue-200/60 px-2 py-0.5 rounded-full">v2.0</span>
        </div>

        <!-- Grouped Navigation -->
        <nav class="flex-1 overflow-y-auto py-4 space-y-6">
          <div>
            <div class="px-6 pb-2 text-[10px] font-bold uppercase tracking-wider text-slate-400">Operasional</div>
            <div class="space-y-0.5">
              ${operasionalLinks}
            </div>
          </div>

          <div>
            <div class="px-6 pb-2 text-[10px] font-bold uppercase tracking-wider text-slate-400">Manajemen PPP</div>
            <div class="space-y-0.5">
              ${manajemenLinks}
            </div>
          </div>

          ${sistemLinks ? `
          <div>
            <div class="px-6 pb-2 text-[10px] font-bold uppercase tracking-wider text-slate-400">Sistem</div>
            <div class="space-y-0.5">
              ${sistemLinks}
            </div>
          </div>
          ` : ''}
        </nav>

        <!-- User Card & Logout Footer -->
        <div class="p-3 border-t border-slate-100 bg-slate-50/60">
          <div class="p-2.5 rounded-xl bg-white border border-slate-200/70 shadow-2xs mb-2 flex items-center justify-between">
            <div class="flex items-center gap-2.5 min-w-0">
              <div class="w-8 h-8 rounded-lg bg-gradient-to-tr from-blue-600 to-indigo-600 text-white flex items-center justify-center font-bold text-xs shrink-0 shadow-xs">
                ${(user.username || 'U')[0].toUpperCase()}
              </div>
              <div class="truncate">
                <p class="text-xs font-bold text-slate-800 truncate leading-tight">${user.full_name || user.username}</p>
                <span class="text-[10px] font-semibold uppercase tracking-wider ${isAdmin ? 'text-purple-600' : 'text-blue-600'}">${roleName}</span>
              </div>
            </div>
          </div>

          <button id="sidebar-logout-btn" class="flex items-center justify-center w-full px-3 py-2 text-xs font-semibold text-rose-600 hover:bg-rose-50 rounded-lg transition-colors gap-2 cursor-pointer">
            <svg xmlns="http://www.w3.org/2000/svg" width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/><polyline points="16 17 21 12 16 7"/><line x1="21" x2="9" y1="12" y2="12"/></svg>
            <span>Keluar Portal</span>
          </button>
        </div>
      </aside>
    `;
  },

  renderHeader(user) {
    const container = document.getElementById('header-container');
    if (!container) return;

    const roleName = (user.roles && user.roles.length > 0) ? user.roles[0] : (user.role || 'Teknisi');
    const isAdmin = Auth.isAdmin();

    container.innerHTML = `
      <header class="hidden lg:flex fixed top-0 right-0 left-64 h-16 glass-header z-30 items-center shadow-2xs">
        <!-- Centered Inner Container aligned with main content max-w-7xl mx-auto -->
        <div class="w-full max-w-7xl mx-auto px-4 sm:px-8 flex items-center justify-between">
          <!-- Left: Router Selector & Ping Badge -->
          <div class="flex items-center gap-3">
            <div class="flex items-center gap-2 px-3 py-1.5 rounded-xl bg-slate-100/90 border border-slate-200/80 text-xs text-slate-700 font-medium">
              <svg xmlns="http://www.w3.org/2000/svg" width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-blue-600"><rect width="20" height="8" x="2" y="14" rx="2"/><path d="M6.01 18h.01"/><path d="M10.01 18h.01"/><path d="M15 10v4"/></svg>
              <span class="text-slate-500">Router:</span>
              <select class="router-select-control bg-transparent border-none text-slate-800 font-semibold outline-none cursor-pointer pr-2 min-w-[180px] max-w-[280px]">
                <option value="">Memuat router...</option>
              </select>
            </div>

            <!-- Streaming Ping Badge -->
            <div id="header-ping-container" class="flex items-center gap-2 px-3 py-1.5 rounded-xl bg-white border border-slate-200/80 text-xs shadow-2xs" title="Latensi Ping MikroTik ke Internet (8.8.8.8)">
              <span class="flex h-2.5 w-2.5 relative">
                <span id="header-ping-ping" class="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
                <span id="header-ping-dot" class="relative inline-flex rounded-full h-2.5 w-2.5 bg-emerald-500"></span>
              </span>
              <span id="header-ping-val" class="font-bold font-mono text-emerald-700 tracking-tight">-- ms</span>
            </div>
          </div>

          <!-- Right: User Identity & Role -->
          <div class="flex items-center gap-3">
            <div class="flex items-center gap-2.5 pl-3 py-1 border-l border-slate-200/70">
              <div class="text-right">
                <span class="text-xs font-bold text-slate-800 leading-tight block">${user.full_name || user.username}</span>
                <span class="text-[10px] font-semibold uppercase tracking-wider ${isAdmin ? 'text-purple-600' : 'text-blue-600'}">${roleName}</span>
              </div>
              <div class="w-8 h-8 rounded-full bg-gradient-to-tr from-blue-600 to-indigo-600 text-white flex items-center justify-center font-bold text-xs uppercase shadow-xs">
                ${(user.username || 'U')[0]}
              </div>
            </div>
          </div>
        </div>
      </header>
    `;
  },

  renderMobileNav(activePage, user, isAdmin) {
    const container = document.getElementById('mobile-nav-container');
    if (!container) return;

    const roleName = (user.roles && user.roles.length > 0) ? user.roles[0] : (user.role || 'Teknisi');

    // 5 Bottom Nav Items (Dashboard, PPP Aktif, PPP Non-Aktif, Log, Monitor)
    const bottomNavItems = [
      {
        title: 'Dashboard',
        href: 'index.html',
        icon: `<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M15 21v-8a1 1 0 0 0-1-1h-4a1 1 0 0 0-1 1v8"/><path d="M3 10a2 2 0 0 1 .709-1.528l7-6a2 2 0 0 1 2.582 0l7 6A2 2 0 0 1 21 10v9a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/></svg>`
      },
      {
        title: 'PPP Aktif',
        href: 'ppp-active.html',
        icon: `<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 20h.01"/><path d="M2 8.82a15 15 0 0 1 20 0"/><path d="M5 12.859a10 10 0 0 1 14 0"/><path d="M8.5 16.429a5 5 0 0 1 7 0"/></svg>`
      },
      {
        title: 'Non-Aktif',
        href: 'ppp-non-active.html',
        icon: `<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 20h.01"/><path d="M8.5 16.429a5 5 0 0 1 7 0"/><path d="M5 12.859a10 10 0 0 1 5.17-2.69"/><path d="M19 12.859a10 10 0 0 0-2.007-1.523"/><path d="m2 2 20 20"/></svg>`
      },
      {
        title: 'Log',
        href: 'mikrotik-logs.html',
        icon: `<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M15 12h-5"/><path d="M15 8h-5"/><path d="M19 17V5a2 2 0 0 0-2-2H4"/><path d="M8 21h12a2 2 0 0 0 2-2v-1a1 1 0 0 0-1-1H11a1 1 0 0 0-1 1v2a1 1 0 0 0 1 1z"/></svg>`
      },
      {
        title: 'Monitor',
        href: 'monitor-interface.html',
        icon: `<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="16" y="16" width="6" height="6" rx="1"/><rect x="2" y="16" width="6" height="6" rx="1"/><rect x="9" y="2" width="6" height="6" rx="1"/><path d="M5 16v-3a1 1 0 0 1 1-1h12a1 1 0 0 1 1 1v3"/><path d="M12 12V8"/></svg>`
      }
    ];

    const bottomNavHtml = bottomNavItems.map(item => {
      const isActive = activePage === item.href || (activePage === '' && item.href === 'index.html');
      return `
        <a href="${item.href}" class="flex flex-col items-center justify-center py-1.5 px-1 rounded-xl transition-all duration-200 group interactive-tap ${
          isActive
            ? 'text-blue-600 font-bold bg-blue-50/90 shadow-2xs'
            : 'text-slate-500 hover:text-slate-900'
        }">
          <div class="relative ${isActive ? 'scale-105' : 'scale-100'} transition-transform">
            ${item.icon}
          </div>
          <span class="text-[10px] mt-0.5 tracking-tight truncate max-w-full font-medium ${isActive ? 'font-bold' : ''}">${item.title}</span>
        </a>
      `;
    }).join('');

    // Drawer links for secondary items & full navigation
    const drawerLinks = this.menus
      .filter(m => !m.adminOnly || isAdmin)
      .map(m => {
        const isActive = activePage === m.href || (activePage === '' && m.href === 'index.html');
        return `
          <a href="${m.href}" class="flex items-center px-4 py-3 gap-3 rounded-xl font-medium text-sm transition-colors ${
            isActive ? 'text-blue-600 bg-blue-50 font-semibold' : 'text-slate-700 hover:bg-slate-100'
          }">
            <span class="${isActive ? 'text-blue-600' : 'text-slate-400'}">${m.icon}</span>
            <span>${m.title}</span>
            ${m.adminOnly ? '<span class="ml-auto text-[9px] uppercase font-bold tracking-wider px-1.5 py-0.5 rounded bg-purple-100 text-purple-700">Admin</span>' : ''}
          </a>
        `;
      }).join('');

    container.innerHTML = `
      <!-- Mobile Top Bar -->
      <header class="lg:hidden fixed top-0 left-0 right-0 h-16 glass-header px-4 z-40 flex items-center justify-between shadow-2xs">
        <div class="flex items-center gap-2 min-w-0">
          <div class="w-8 h-8 rounded-lg bg-gradient-to-tr from-blue-600 to-indigo-600 text-white flex items-center justify-center shrink-0 shadow-xs">
            <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="20" height="8" x="2" y="14" rx="2"/><path d="M6.01 18h.01"/><path d="M10.01 18h.01"/><path d="M15 10v4"/><path d="M17.84 7.17a4 4 0 0 0-5.66 0"/><path d="M20.66 4.34a8 8 0 0 0-11.31 0"/></svg>
          </div>
          <span class="font-bold text-base text-slate-900 tracking-tight">GNet Dial</span>
        </div>

        <!-- Mobile Center/Right: Router Select & Ping -->
        <div class="flex items-center gap-2">
          <!-- Mobile Ping Badge -->
          <div class="flex items-center gap-1.5 px-2 py-1 rounded-lg bg-slate-100 border border-slate-200 text-xs font-mono shadow-2xs" title="Latensi Ping">
            <span class="flex h-2 w-2 relative">
              <span id="mobile-header-ping-dot" class="relative inline-flex rounded-full h-2 w-2 bg-emerald-500"></span>
            </span>
            <span id="mobile-header-ping-val" class="font-bold text-emerald-700">-- ms</span>
          </div>

          <!-- Drawer Trigger Button -->
          <button id="mobile-menu-open-btn" class="p-2 text-slate-600 hover:text-slate-900 rounded-lg hover:bg-slate-100 transition-colors cursor-pointer" aria-label="Menu">
            <svg xmlns="http://www.w3.org/2000/svg" width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="4" x2="20" y1="12" y2="12"/><line x1="4" x2="20" y1="6" y2="6"/><line x1="4" x2="20" y1="18" y2="18"/></svg>
          </button>
        </div>
      </header>

      <!-- Mobile Bottom Navigation Bar (5 Core Tabs) -->
      <nav class="lg:hidden fixed bottom-0 left-0 right-0 z-40 bottom-nav-glass px-2 pt-1 pb-safe shadow-lg">
        <div class="grid grid-cols-5 gap-1 items-center justify-around">
          ${bottomNavHtml}
        </div>
      </nav>

      <!-- Mobile Slide-over Drawer -->
      <div id="mobile-drawer" class="lg:hidden fixed inset-0 z-50 hidden">
        <div id="mobile-drawer-backdrop" class="fixed inset-0 bg-slate-900/60 backdrop-blur-xs transition-opacity"></div>
        <div class="fixed top-0 bottom-0 right-0 w-80 max-w-[85vw] bg-white flex flex-col shadow-2xl z-10 transform translate-x-full transition-transform duration-300" id="mobile-drawer-panel">
          <!-- Drawer Header -->
          <div class="flex items-center justify-between h-16 px-5 border-b border-slate-100 bg-slate-50/50">
            <div class="flex items-center gap-2.5">
              <div class="w-8 h-8 rounded-full bg-gradient-to-tr from-blue-600 to-indigo-600 text-white flex items-center justify-center font-bold text-xs">
                ${(user.username || 'U')[0].toUpperCase()}
              </div>
              <div>
                <span class="text-xs font-bold text-slate-800 block">${user.full_name || user.username}</span>
                <span class="text-[10px] font-semibold uppercase tracking-wider text-slate-500">${roleName}</span>
              </div>
            </div>
            <button id="mobile-menu-close-btn" class="p-2 text-slate-400 hover:text-slate-600 rounded-lg hover:bg-slate-100 cursor-pointer">
              <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 6 6 18"/><path d="m6 6 12 12"/></svg>
            </button>
          </div>

          <!-- Router Selector in Drawer -->
          <div class="p-4 border-b border-slate-100 bg-slate-50/80">
            <label class="block text-[10px] font-bold text-slate-500 uppercase tracking-wider mb-1.5">Router Aktif</label>
            <select class="router-select-control w-full bg-white border border-slate-300 text-slate-800 text-xs rounded-xl px-3 py-2.5 font-semibold outline-none shadow-2xs">
              <option value="">Memuat router...</option>
            </select>
          </div>

          <!-- Menu Links -->
          <div class="px-3 py-2 flex-1 overflow-y-auto space-y-1">
            <div class="px-3 pt-2 pb-1 text-[10px] font-bold uppercase tracking-wider text-slate-400">Seluruh Menu</div>
            ${drawerLinks}
          </div>

          <!-- Drawer Footer Logout -->
          <div class="p-4 border-t border-slate-100 bg-slate-50">
            <button id="mobile-logout-btn" class="w-full flex items-center justify-center gap-2 py-2.5 px-4 bg-rose-50 hover:bg-rose-100 text-rose-600 rounded-xl text-xs font-bold transition-colors cursor-pointer">
              <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/><polyline points="16 17 21 12 16 7"/><line x1="21" x2="9" y1="12" y2="12"/></svg>
              <span>Keluar Portal</span>
            </button>
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
        panel.classList.remove('translate-x-full');
      });
    };

    const closeDrawer = () => {
      if (!drawer) return;
      panel.classList.add('translate-x-full');
      setTimeout(() => drawer.classList.add('hidden'), 300);
    };

    if (openBtn) openBtn.onclick = openDrawer;
    if (closeBtn) closeBtn.onclick = closeDrawer;
    if (backdrop) backdrop.onclick = closeDrawer;
  }
};

window.Layout = Layout;
