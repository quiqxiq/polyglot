// DEVIASI: layout template strings, icons, and navigation markup exceed 500 lines
/**
 * Polyglot Web Dial Layout Engine
 * Merender navigasi Sidebar, Centered Header, Mobile Top Bar, dan Bottom
 * Navigation Bar dari registry menu yang sudah difilter permission (DialMenu).
 *
 * Tidak ada lagi flag adminOnly: setiap node, submenu, dan aksi mengikuti
 * permission efektif akun. Router selector dan ping hanya dirender bila akun
 * memiliki device:read / device:diagnostic:exec.
 */

const Layout = {
  pingStreamController: null,
  currentPingDeviceId: null,

  esc(value) {
    return String(value === null || value === undefined ? '' : value)
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#039;');
  },

  t(key, fallback) {
    if (typeof I18n !== 'undefined') return I18n.t(key);
    return fallback || key;
  },

  label(node) {
    return node.labelKey ? this.t(node.labelKey, node.title) : (node.title || node.id);
  },

  canUseRouterSelector() {
    return typeof Auth !== 'undefined' && (Auth.can('dial:device:read') || Auth.can('device:read'));
  },

  canUseHeaderPing() {
    return typeof Auth !== 'undefined' && (
      Auth.can('dial:dashboard:ping') ||
      Auth.can('dial:ppp-active:ping') ||
      Auth.can('dial:monitor:diagnostic') ||
      Auth.can('device:diagnostic:exec')
    );
  },

  normalizePage(href) {
    return String(href || 'index.html').split('?')[0];
  },

  isNodeActive(node, activePage) {
    if (!node || !node.href) return false;
    const active = this.normalizePage(activePage);
    const target = this.normalizePage(node.href);
    if (active !== target) return false;
    const activeHash = String(activePage || '').split('#')[1] || '';
    const nodeHash = String(node.href).split('#')[1] || '';
    if (nodeHash) return activeHash === nodeHash;
    return !activeHash || !(node.children || []).some((child) => child.href && String(child.href).includes('#'));
  },

  async init(activePageHref = 'index.html') {
    const user = Auth.getUser() || { username: 'Pengguna', role: 'user', permissions: [] };

    if (typeof I18n !== 'undefined' && I18n.applyDOM) I18n.applyDOM();

    this.renderSidebar(activePageHref, user);
    this.renderHeader(user);
    this.renderMobileNav(activePageHref, user);
    this.bindEvents();

    if (typeof I18n !== 'undefined' && I18n.applyDOM) I18n.applyDOM();

    // Router list and header ping are API calls; only start them when the
    // account actually holds the matching backend permission.
    if (this.canUseRouterSelector()) {
      await RouterSelector.init();

      RouterSelector.onDeviceChange((deviceId) => {
        if (!deviceId || !this.canUseHeaderPing()) return;
        this.startHeaderPingStream(deviceId);
      });

      const initialDeviceId = RouterSelector.getSelectedDeviceId();
      if (initialDeviceId && this.canUseHeaderPing()) {
        this.startHeaderPingStream(initialDeviceId);
      }
    }

    window.addEventListener('beforeunload', () => {
      this.cleanupPingStream();
    });
  },

  rerender(activePageHref = 'index.html') {
    const user = Auth.getUser() || { username: 'Pengguna', role: 'user', permissions: [] };
    this.renderSidebar(activePageHref, user);
    this.renderHeader(user);
    this.renderMobileNav(activePageHref, user);
    this.bindEvents();
    if (this.canUseRouterSelector() && typeof RouterSelector !== 'undefined' && RouterSelector.devices.length > 0) {
      RouterSelector.renderDropdowns();
    }
  },

  cleanupPingStream() {
    if (this.pingStreamController) {
      this.pingStreamController.abort();
      this.pingStreamController = null;
    }
    this.currentPingDeviceId = null;
  },

  startHeaderPingStream(deviceId) {
    if (!deviceId || !this.canUseHeaderPing()) return;
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

      if (status === 'timeout' || status === 'error' || status === 'high') {
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
        // Never spin on an authorization failure.
        if (err && (err.permissionDenied || err.status === 403)) return;
        setTimeout(() => {
          if (!signal.aborted) this.startHeaderPingStream(deviceId);
        }, 5000);
      },
      signal
    );
  },

  renderSidebar(activePage, user) {
    const container = document.getElementById('sidebar-container');
    if (!container || typeof DialMenu === 'undefined') return;

    const tree = DialMenu.visibleTree(Auth.permissionChecker());
    const navHtml = tree.map((node) => this.renderSidebarNode(node, 0, activePage)).join('');

    const roleName = Auth.primaryRole(user);
    const isAdmin = Auth.isAdmin();
    const logoutText = this.t('logout', 'Keluar Portal');
    const displayName = this.esc(user.fullName || user.full_name || user.username || 'Pengguna');
    const initial = this.esc((user.username || 'U')[0].toUpperCase());

    container.innerHTML = `
      <aside class="hidden lg:flex fixed top-0 left-0 h-screen w-64 bg-white border-r border-slate-200/80 z-40 flex-col shadow-xs select-none">
        <div class="flex items-center justify-between h-16 px-6 border-b border-slate-100 bg-white">
          <div class="flex items-center gap-2.5">
            <div class="w-9 h-9 rounded-xl bg-gradient-to-tr from-blue-600 to-indigo-600 text-white flex items-center justify-center shadow-md shadow-blue-500/20">
              <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="20" height="8" x="2" y="14" rx="2"/><path d="M6.01 18h.01"/><path d="M10.01 18h.01"/><path d="M15 10v4"/><path d="M17.84 7.17a4 4 0 0 0-5.66 0"/><path d="M20.66 4.34a8 8 0 0 0-11.31 0"/></svg>
            </div>
            <div>
              <span class="font-bold text-base text-slate-900 tracking-tight leading-none block">GNet Dial</span>
              <span class="text-[10px] font-semibold text-slate-400 uppercase tracking-wider block mt-0.5">${this.t('group_operations', 'Operasional')}</span>
            </div>
          </div>
          <span class="text-[10px] font-bold bg-blue-50 text-blue-700 border border-blue-200/60 px-2 py-0.5 rounded-full">v2.1</span>
        </div>

        <nav class="flex-1 overflow-y-auto py-4 space-y-6">
          ${navHtml}
        </nav>

        <div class="p-3 border-t border-slate-100 bg-slate-50/60">
          <div class="p-2.5 rounded-xl bg-white border border-slate-200/70 shadow-2xs mb-2 flex items-center justify-between">
            <div class="flex items-center gap-2.5 min-w-0">
              <div class="w-8 h-8 rounded-lg bg-gradient-to-tr from-blue-600 to-indigo-600 text-white flex items-center justify-center font-bold text-xs shrink-0 shadow-xs">
                ${initial}
              </div>
              <div class="truncate">
                <p class="text-xs font-bold text-slate-800 truncate leading-tight">${displayName}</p>
                <span class="text-[10px] font-semibold uppercase tracking-wider ${isAdmin ? 'text-purple-600' : 'text-blue-600'}">${this.esc(roleName)}</span>
              </div>
            </div>
          </div>

          <button id="sidebar-logout-btn" class="flex items-center justify-center w-full px-3 py-2 text-xs font-semibold text-rose-600 hover:bg-rose-50 rounded-lg transition-colors gap-2 cursor-pointer">
            <svg xmlns="http://www.w3.org/2000/svg" width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/><polyline points="16 17 21 12 16 7"/><line x1="21" x2="9" y1="12" y2="12"/></svg>
            <span>${logoutText}</span>
          </button>
        </div>
      </aside>
    `;
  },

  renderSidebarNode(node, depth, activePage) {
    if (node.type === 'group') {
      const children = (node.children || [])
        .map((child) => this.renderSidebarNode(child, depth + 1, activePage))
        .join('');
      if (!children) return '';

      const headingClass = depth === 0
        ? 'px-6 pb-2 text-[10px] font-bold uppercase tracking-wider text-slate-400'
        : 'px-6 pt-3 pb-1.5 text-[10px] font-bold uppercase tracking-wider text-slate-400/80';
      return `
        <div>
          <div class="${headingClass}">${this.label(node)}</div>
          <div class="space-y-0.5">${children}</div>
        </div>
      `;
    }

    const isActive = this.isNodeActive(node, activePage);
    const isTab = node.type === 'tab';
    const padding = isTab ? 'pl-10 pr-3.5 py-2 text-xs' : 'px-3.5 py-2.5 mx-3 text-sm';
    const icon = isTab ? '' : `<span class="${isActive ? 'text-blue-600' : 'text-slate-400 group-hover:text-slate-600'} transition-colors">${DialMenu.icon(node.icon)}</span>`;
    const link = `
      <a href="${node.href}" class="flex items-center ${padding} rounded-xl gap-3 font-medium transition-all duration-200 group ${
        isActive
          ? 'text-blue-600 bg-blue-50/90 font-semibold shadow-xs'
          : 'text-slate-600 hover:bg-slate-50 hover:text-slate-900'
      }">
        ${icon}
        <span>${this.label(node)}</span>
      </a>
    `;

    if (node.type === 'page' && node.children && node.children.length > 0) {
      const tabs = node.children
        .map((child) => this.renderSidebarNode(child, depth + 1, activePage))
        .join('');
      return `${link}<div class="space-y-0.5">${tabs}</div>`;
    }

    return link;
  },

  renderHeader(user) {
    const container = document.getElementById('header-container');
    if (!container) return;

    const roleName = Auth.primaryRole(user);
    const isAdmin = Auth.isAdmin();
    const currentLang = typeof I18n !== 'undefined' ? I18n.getLanguage() : 'id';
    const showRouter = this.canUseRouterSelector();
    const showPing = this.canUseHeaderPing();
    const displayName = this.esc(user.fullName || user.full_name || user.username || 'Pengguna');
    const initial = this.esc((user.username || 'U')[0].toUpperCase());

    const routerHtml = showRouter ? `
      <div class="flex items-center gap-2 px-3 py-1.5 rounded-xl bg-slate-100/90 border border-slate-200/80 text-xs text-slate-700 font-medium">
        <svg xmlns="http://www.w3.org/2000/svg" width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-blue-600"><rect width="20" height="8" x="2" y="14" rx="2"/><path d="M6.01 18h.01"/><path d="M10.01 18h.01"/><path d="M15 10v4"/></svg>
        <span class="text-slate-500">Router:</span>
        <select class="router-select-control bg-transparent border-none text-slate-800 font-semibold outline-none cursor-pointer pr-2 min-w-[180px] max-w-[280px]">
          <option value="">${this.t('select_router', 'Memuat router...')}</option>
        </select>
      </div>
    ` : '';

    const pingHtml = showPing ? `
      <div id="header-ping-container" class="flex items-center gap-2 px-3 py-1.5 rounded-xl bg-white border border-slate-200/80 text-xs shadow-2xs" title="Latensi Ping MikroTik ke Internet (8.8.8.8)">
        <span class="flex h-2.5 w-2.5 relative">
          <span id="header-ping-ping" class="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
          <span id="header-ping-dot" class="relative inline-flex rounded-full h-2.5 w-2.5 bg-emerald-500"></span>
        </span>
        <span id="header-ping-val" class="font-bold font-mono text-emerald-700 tracking-tight">-- ms</span>
      </div>
    ` : '';

    container.innerHTML = `
      <header class="hidden lg:flex fixed top-0 right-0 left-64 h-16 glass-header z-30 items-center shadow-2xs">
        <div class="w-full max-w-7xl mx-auto px-4 sm:px-8 flex items-center justify-between">
          <div class="flex items-center gap-3">
            ${routerHtml}
            ${pingHtml}
          </div>

          <div class="flex items-center gap-3">
            <button onclick="I18n.toggleLanguage()" class="flex items-center gap-1.5 px-2.5 py-1.5 rounded-xl bg-slate-100/90 hover:bg-slate-200/80 border border-slate-200/80 text-xs font-semibold text-slate-700 transition-colors cursor-pointer" title="Switch Language (ID / EN)">
              <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-blue-600"><circle cx="12" cy="12" r="10"/><path d="M12 2a14.5 14.5 0 0 0 0 20 14.5 14.5 0 0 0 0-20"/><path d="M2 12h20"/></svg>
              <span class="uppercase font-bold">${currentLang}</span>
            </button>

            <div class="flex items-center gap-2.5 pl-3 py-1 border-l border-slate-200/70">
              <div class="text-right">
                <span class="text-xs font-bold text-slate-800 leading-tight block">${displayName}</span>
                <span class="text-[10px] font-semibold uppercase tracking-wider ${isAdmin ? 'text-purple-600' : 'text-blue-600'}">${this.esc(roleName)}</span>
              </div>
              <div class="w-8 h-8 rounded-full bg-gradient-to-tr from-blue-600 to-indigo-600 text-white flex items-center justify-center font-bold text-xs uppercase shadow-xs">
                ${initial}
              </div>
            </div>
          </div>
        </div>
      </header>
    `;
  },

  renderDrawerNodes(nodes, depth = 0) {
    return (nodes || []).map((node) => {
      if (node.type === 'group') {
        const children = this.renderDrawerNodes(node.children, depth + 1);
        if (!children) return '';
        return `
          <div class="${depth > 0 ? 'pl-2' : 'pt-3'}">
            <div class="px-4 pb-1 text-[10px] font-bold uppercase tracking-wider text-slate-400">${this.label(node)}</div>
            ${children}
          </div>
        `;
      }

      const isTab = node.type === 'tab';
      const active = this.isNodeActive(node, window.location.pathname.split('/').pop() || 'index.html');
      return `
        <a href="${node.href}" class="flex items-center ${isTab ? 'pl-10 pr-4 py-2 text-xs' : 'px-4 py-3 text-sm'} gap-3 rounded-xl font-medium transition-colors ${
          active ? 'text-blue-600 bg-blue-50 font-semibold' : 'text-slate-700 hover:bg-slate-100'
        }">
          ${isTab ? '' : `<span class="${active ? 'text-blue-600' : 'text-slate-400'}">${DialMenu.icon(node.icon)}</span>`}
          <span>${this.label(node)}</span>
        </a>
      `;
    }).join('');
  },

  renderMobileNav(activePage, user) {
    const container = document.getElementById('mobile-nav-container');
    if (!container || typeof DialMenu === 'undefined') return;

    const roleName = Auth.primaryRole(user);
    const tree = DialMenu.visibleTree(Auth.permissionChecker());
    const items = DialMenu.bottomNavItems(Auth.permissionChecker());
    const showRouter = this.canUseRouterSelector();
    const showPing = this.canUseHeaderPing();
    const currentLang = typeof I18n !== 'undefined' ? I18n.getLanguage() : 'id';
    const logoutText = this.t('logout', 'Keluar Portal');
    const initial = this.esc((user.username || 'U')[0].toUpperCase());
    const displayName = this.esc(user.fullName || user.full_name || user.username || 'Pengguna');

    const bottomNavHtml = items.map((item) => {
      const isActive = this.isNodeActive(item, activePage) ||
        (activePage === '' && item.href === 'index.html');
      return `
        <a href="${item.href}" class="flex flex-col items-center justify-center py-1.5 px-1 rounded-xl transition-all duration-200 group interactive-tap ${
          isActive
            ? 'text-blue-600 font-bold bg-blue-50/90 shadow-2xs'
            : 'text-slate-500 hover:text-slate-900'
        }">
          <div class="relative ${isActive ? 'scale-105' : 'scale-100'} transition-transform">
            ${DialMenu.icon(item.icon)}
          </div>
          <span class="text-[10px] mt-0.5 tracking-tight truncate max-w-full font-medium ${isActive ? 'font-bold' : ''}">${this.label(item)}</span>
        </a>
      `;
    }).join('');

    const bottomNav = items.length > 0 ? `
      <nav class="lg:hidden fixed bottom-0 left-0 right-0 z-40 bottom-nav-glass px-2 pt-1 pb-safe shadow-lg">
        <div class="grid gap-1 items-center justify-around" style="grid-template-columns: repeat(${items.length}, minmax(0, 1fr));">
          ${bottomNavHtml}
        </div>
      </nav>
    ` : '';

    const drawerLinks = this.renderDrawerNodes(tree);

    container.innerHTML = `
      <header class="lg:hidden fixed top-0 left-0 right-0 h-16 glass-header px-4 z-40 flex items-center justify-between shadow-2xs">
        <div class="flex items-center gap-2 min-w-0">
          <div class="w-8 h-8 rounded-lg bg-gradient-to-tr from-blue-600 to-indigo-600 text-white flex items-center justify-center shrink-0 shadow-xs">
            <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="20" height="8" x="2" y="14" rx="2"/><path d="M6.01 18h.01"/><path d="M10.01 18h.01"/><path d="M15 10v4"/><path d="M17.84 7.17a4 4 0 0 0-5.66 0"/><path d="M20.66 4.34a8 8 0 0 0-11.31 0"/></svg>
          </div>
          <span class="font-bold text-base text-slate-900 tracking-tight">GNet Dial</span>
        </div>

        <div class="flex items-center gap-2">
          ${showPing ? `
          <div class="flex items-center gap-1.5 px-2 py-1 rounded-lg bg-slate-100 border border-slate-200 text-xs font-mono shadow-2xs" title="Latensi Ping">
            <span class="flex h-2 w-2 relative">
              <span id="mobile-header-ping-dot" class="relative inline-flex rounded-full h-2 w-2 bg-emerald-500"></span>
            </span>
            <span id="mobile-header-ping-val" class="font-bold text-emerald-700">-- ms</span>
          </div>
          ` : ''}

          <button onclick="I18n.toggleLanguage()" class="px-2 py-1 rounded-lg bg-slate-100 border border-slate-200 text-xs font-bold text-slate-700 uppercase" title="Ganti Bahasa">
            ${currentLang}
          </button>

          <button id="mobile-menu-open-btn" class="p-2 text-slate-600 hover:text-slate-900 rounded-lg hover:bg-slate-100 transition-colors cursor-pointer" aria-label="Menu">
            <svg xmlns="http://www.w3.org/2000/svg" width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="4" x2="20" y1="12" y2="12"/><line x1="4" x2="20" y1="6" y2="6"/><line x1="4" x2="20" y1="18" y2="18"/></svg>
          </button>
        </div>
      </header>

      ${bottomNav}

      <div id="mobile-drawer" class="lg:hidden fixed inset-0 z-50 hidden">
        <div id="mobile-drawer-backdrop" class="fixed inset-0 bg-slate-900/60 backdrop-blur-xs transition-opacity"></div>
        <div class="fixed top-0 bottom-0 right-0 w-80 max-w-[85vw] bg-white flex flex-col shadow-2xl z-10 transform translate-x-full transition-transform duration-300" id="mobile-drawer-panel">
          <div class="flex items-center justify-between h-16 px-5 border-b border-slate-100 bg-slate-50/50">
            <div class="flex items-center gap-2.5">
              <div class="w-8 h-8 rounded-full bg-gradient-to-tr from-blue-600 to-indigo-600 text-white flex items-center justify-center font-bold text-xs">
                ${initial}
              </div>
              <div>
                <span class="text-xs font-bold text-slate-800 block">${displayName}</span>
                <span class="text-[10px] font-semibold uppercase tracking-wider text-slate-500">${this.esc(roleName)}</span>
              </div>
            </div>
            <button id="mobile-menu-close-btn" class="p-2 text-slate-400 hover:text-slate-600 rounded-lg hover:bg-slate-100 cursor-pointer">
              <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 6 6 18"/><path d="m6 6 12 12"/></svg>
            </button>
          </div>

          ${showRouter ? `
          <div class="p-4 border-b border-slate-100 bg-slate-50/80">
            <label class="block text-[10px] font-bold text-slate-500 uppercase tracking-wider mb-1.5">${this.t('select_router', 'Router Aktif')}</label>
            <select class="router-select-control w-full bg-white border border-slate-300 text-slate-800 text-xs rounded-xl px-3 py-2.5 font-semibold outline-none shadow-2xs">
              <option value="">${this.t('select_router', 'Memuat router...')}</option>
            </select>
          </div>
          ` : ''}

          <div class="px-3 py-2 flex-1 overflow-y-auto space-y-1">
            <div class="px-3 pt-2 pb-1 text-[10px] font-bold uppercase tracking-wider text-slate-400">${currentLang === 'en' ? 'All Menus' : 'Seluruh Menu'}</div>
            ${drawerLinks}
          </div>

          <div class="p-4 border-t border-slate-100 bg-slate-50">
            <button id="mobile-logout-btn" class="w-full flex items-center justify-center gap-2 py-2.5 px-4 bg-rose-50 hover:bg-rose-100 text-rose-600 rounded-xl text-xs font-bold transition-colors cursor-pointer">
              <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/><polyline points="16 17 21 12 16 7"/><line x1="21" x2="9" y1="12" y2="12"/></svg>
              <span>${logoutText}</span>
            </button>
          </div>
        </div>
      </div>
    `;
  },

  bindEvents() {
    const sideLogout = document.getElementById('sidebar-logout-btn');
    if (sideLogout) sideLogout.onclick = () => Auth.logout();

    const mobLogout = document.getElementById('mobile-logout-btn');
    if (mobLogout) mobLogout.onclick = () => Auth.logout();

    const openBtn = document.getElementById('mobile-menu-open-btn');
    const closeBtn = document.getElementById('mobile-menu-close-btn');
    const drawer = document.getElementById('mobile-drawer');
    const backdrop = document.getElementById('mobile-drawer-backdrop');
    const panel = document.getElementById('mobile-drawer-panel');

    const openDrawer = () => {
      if (!drawer || !panel) return;
      drawer.classList.remove('hidden');
      requestAnimationFrame(() => {
        panel.classList.remove('translate-x-full');
      });
    };

    const closeDrawer = () => {
      if (!drawer || !panel) return;
      panel.classList.add('translate-x-full');
      setTimeout(() => drawer.classList.add('hidden'), 300);
    };

    if (openBtn) openBtn.onclick = openDrawer;
    if (closeBtn) closeBtn.onclick = closeDrawer;
    if (backdrop) backdrop.onclick = closeDrawer;
  }
};

window.Layout = Layout;
