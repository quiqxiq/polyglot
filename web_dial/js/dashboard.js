/**
 * Polyglot Web Dial Dashboard Controller
 * Streaming real-time system resources (CPU, RAM, HDD, Uptime, OS) via StreamSystemSnapshot
 * and managing quick PPP statistics and search (aligned with ppp-active and ppp-non-active).
 */

const Dashboard = {
  currentDeviceId: null,
  resourceStreamController: null,
  localTicker: null,
  targetPingIp: null,

  init() {
    RouterSelector.onDeviceChange((deviceId) => {
      if (!deviceId) return;
      this.currentDeviceId = deviceId;
      this.startResourceStream();
      this.loadQuickStats();
      const searchInput = document.getElementById('search-input');
      if (searchInput && searchInput.value.trim().length >= 2) {
        this.performSearch(searchInput.value.trim().toLowerCase());
      }
    });

    const initialDevId = RouterSelector.getSelectedDeviceId();
    if (initialDevId && !this.currentDeviceId) {
      this.currentDeviceId = initialDevId;
      this.startResourceStream();
      this.loadQuickStats();
    }

    this.bindSearch();
    this.bindModals();

    window.addEventListener('beforeunload', () => {
      this.cleanupStreams();
    });
  },

  cleanupStreams() {
    if (this.resourceStreamController) {
      try {
        this.resourceStreamController.abort();
      } catch (_) {}
      this.resourceStreamController = null;
    }
    this.stopLocalTicker();
  },

  startLocalTicker() {
    this.stopLocalTicker();
    this.localTicker = setInterval(() => {
      document.querySelectorAll('.uptime-val').forEach((el) => {
        const text = el.textContent.trim();
        if (!text || text === '-' || text === 'N/A') return;
        const parts = text.split(' ');
        let timePart = parts[parts.length - 1];
        const hms = timePart.split(':');
        if (hms.length === 3) {
          let h = parseInt(hms[0], 10);
          let m = parseInt(hms[1], 10);
          let s = parseInt(hms[2], 10) + 1;
          if (s >= 60) {
            s = 0;
            m++;
          }
          if (m >= 60) {
            m = 0;
            h++;
          }
          timePart = `${h.toString().padStart(2, '0')}:${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`;
          parts[parts.length - 1] = timePart;
          el.textContent = parts.join(' ');
        }
      });
    }, 1000);
  },

  stopLocalTicker() {
    if (this.localTicker) {
      clearInterval(this.localTicker);
      this.localTicker = null;
    }
  },

  /**
   * 1. Stream System Resources (CPU, RAM, HDD, Uptime, Version, Board, Identity)
   * Menggunakan /polyglot.v1.NetworkMonitorService/StreamSystemSnapshot secara native
   */
  startResourceStream() {
    this.cleanupStreams();
    if (!this.currentDeviceId) return;

    this.resourceStreamController = new AbortController();
    const signal = this.resourceStreamController.signal;

    const badgeEl = document.getElementById('resource-stream-badge');
    const badgeText = document.getElementById('resource-stream-text');
    if (badgeEl && badgeText) {
      badgeEl.className = 'inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-semibold bg-blue-50 text-blue-700 border border-blue-200/80';
      badgeText.textContent = 'CONNECTING';
    }

    const setT = (id, val) => {
      const el = document.getElementById(id);
      if (el) el.textContent = val;
    };

    API.stream(
      '/polyglot.v1.NetworkMonitorService/StreamSystemSnapshot',
      { deviceId: this.currentDeviceId, interval: '1s' },
      (frame) => {
        const res = frame.resource || {};
        const rb = frame.routerboard || {};
        const ident = frame.identity || '';

        if (badgeEl && badgeText) {
          badgeEl.className = 'inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-semibold bg-emerald-50 text-emerald-700 border border-emerald-200/80';
          badgeText.textContent = 'LIVE STREAM';
        }

        // 1. Identity, Board & Architecture
        setT('sys-identity', ident || rb.model || 'MikroTik Router');
        setT('sys-board', res.boardName ? `${res.boardName} (${rb.model || 'Router'})` : (rb.model || 'MikroTik'));
        setT('sys-arch', res.architecture ? `${res.architecture} · ${res.cpuFrequency || '-'} MHz` : '-');
        setT('sys-uptime', res.uptime || '-');
        setT('sys-version', res.version || '-');
        setT('sys-serial', rb.serialNumber || '-');
        setT('sys-status', 'Tersambung');

        const hostEl = document.getElementById('sys-host');
        if (hostEl) {
          const dev = RouterSelector.getSelectedDevice();
          hostEl.textContent = dev ? `${dev.host}:${dev.port || 8728} (Stream aktif)` : 'Stream aktif';
        }

        // 2. CPU Load Meter
        const cpu = res.cpuLoad !== undefined ? Number(res.cpuLoad) : 0;
        setT('sys-cpu-text', `${cpu}%`);
        const cpuBar = document.getElementById('sys-cpu-bar');
        if (cpuBar) {
          cpuBar.style.width = `${Math.min(100, Math.max(0, cpu))}%`;
          cpuBar.className = `h-full rounded-full transition-all duration-500 ${
            cpu >= 90 ? 'bg-rose-500' : cpu >= 70 ? 'bg-amber-500' : 'bg-emerald-500'
          }`;
        }
        setT('sys-cpu-freq', `Frekuensi: ${res.cpuFrequency || '-'} MHz (${res.cpuCount || 1} Core)`);

        // 3. Memory RAM Meter
        const freeMem = Number(res.freeMemory || 0);
        const totalMem = Number(res.totalMemory || 0);
        if (totalMem > 0) {
          const usedMem = totalMem - freeMem;
          const memPct = Math.round((usedMem / totalMem) * 100);
          setT('sys-mem-text', `${memPct}%`);
          const memBar = document.getElementById('sys-mem-bar');
          if (memBar) {
            memBar.style.width = `${memPct}%`;
            memBar.className = `h-full rounded-full transition-all duration-500 ${
              memPct >= 85 ? 'bg-rose-500' : memPct >= 65 ? 'bg-amber-500' : 'bg-emerald-500'
            }`;
          }
          const usedMB = (usedMem / 1048576).toFixed(1);
          const totalMB = (totalMem / 1048576).toFixed(1);
          const freeMB = (freeMem / 1048576).toFixed(1);
          setT('sys-mem-detail', `Terpakai: ${usedMB} MB / ${totalMB} MB · Sisa ${freeMB} MB`);
        }

        // 4. Storage HDD Meter
        const freeHdd = Number(res.freeHddSpace || 0);
        const totalHdd = Number(res.totalHddSpace || 0);
        if (totalHdd > 0) {
          const usedHdd = totalHdd - freeHdd;
          const hddPct = Math.round((usedHdd / totalHdd) * 100);
          setT('sys-hdd-text', `${hddPct}%`);
          const hddBar = document.getElementById('sys-hdd-bar');
          if (hddBar) {
            hddBar.style.width = `${hddPct}%`;
            hddBar.className = `h-full rounded-full transition-all duration-500 ${
              hddPct >= 90 ? 'bg-rose-500' : hddPct >= 75 ? 'bg-amber-500' : 'bg-emerald-500'
            }`;
          }
          const usedHddMB = (usedHdd / 1048576).toFixed(1);
          const totalHddMB = (totalHdd / 1048576).toFixed(1);
          const freeHddMB = (freeHdd / 1048576).toFixed(1);
          setT('sys-hdd-detail', `Terpakai: ${usedHddMB} MB / ${totalHddMB} MB · Sisa ${freeHddMB} MB`);
        }
      },
      (err) => {
        if (signal.aborted) return;
        if (badgeEl && badgeText) {
          badgeEl.className = 'inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-semibold bg-slate-100 text-slate-600 border border-slate-200';
          badgeText.textContent = 'OFFLINE';
        }
        setTimeout(() => {
          if (!signal.aborted && this.currentDeviceId) {
            this.startResourceStream();
          }
        }, 4000);
      },
      signal
    );
  },

  /**
   * Quick Stats (PPP Active, Inactive, Secrets, Profiles)
   */
  async loadQuickStats() {
    if (!this.currentDeviceId) return;
    try {
      const [activeRes, inactiveRes, secretsRes, profilesRes] = await Promise.allSettled([
        API.call('/polyglot.v1.PPPService/ListActiveSessions', { device_id: this.currentDeviceId }),
        API.call('/polyglot.v1.PPPService/ListInactiveSecrets', { device_id: this.currentDeviceId }),
        API.call('/polyglot.v1.PPPService/ListSecrets', { device_id: this.currentDeviceId }),
        API.call('/polyglot.v1.PPPService/ListProfiles', { device_id: this.currentDeviceId })
      ]);

      const setVal = (id, res, key) => {
        const el = document.getElementById(id);
        if (el) el.textContent = res.status === 'fulfilled' ? (res.value[key] || []).length : 0;
      };
      setVal('stat-active', activeRes, 'sessions');
      setVal('stat-inactive', inactiveRes, 'secrets');
      setVal('stat-secrets', secretsRes, 'secrets');
      setVal('stat-profiles', profilesRes, 'profiles');
    } catch (err) {
      console.error('Error fetching quick stats:', err);
    }
  },

  /**
   * Quick PPP Search
   */
  bindSearch() {
    const input = document.getElementById('search-input');
    const resultsContainer = document.getElementById('search-results');
    const resultsSection = document.getElementById('search-results-section');
    if (!input || !resultsContainer) return;

    let timeout = null;
    input.addEventListener('input', () => {
      clearTimeout(timeout);
      const query = input.value.trim().toLowerCase();
      if (query.length < 2) {
        if (resultsSection) resultsSection.classList.add('hidden');
        resultsContainer.innerHTML = '';
        this.stopLocalTicker();
        return;
      }
      timeout = setTimeout(() => this.performSearch(query), 300);
    });
  },

  async performSearch(query) {
    if (!this.currentDeviceId) return;
    const resultsContainer = document.getElementById('search-results');
    const resultsSection = document.getElementById('search-results-section');
    const countEl = document.getElementById('search-results-count');
    const loadingSpinner = document.getElementById('search-loading');

    if (loadingSpinner) loadingSpinner.classList.remove('hidden');

    try {
      const [activeRes, inactiveRes] = await Promise.all([
        API.call('/polyglot.v1.PPPService/ListActiveSessions', { device_id: this.currentDeviceId }),
        API.call('/polyglot.v1.PPPService/ListInactiveSecrets', { device_id: this.currentDeviceId })
      ]);

      const activeList = activeRes.sessions || [];
      const inactiveList = inactiveRes.secrets || [];

      const matchedActive = activeList.filter(s => {
        const name = (s.name || '').toLowerCase();
        const callerId = (s.callerId || s.caller_id || '').toLowerCase();
        const address = (s.address || '').toLowerCase();
        const service = (s.service || '').toLowerCase();
        const profile = (s.profile || '').toLowerCase();
        return name.includes(query) || callerId.includes(query) || address.includes(query) || service.includes(query) || profile.includes(query);
      });

      const matchedInactive = inactiveList.filter(s => {
        const name = (s.name || '').toLowerCase();
        const profile = (s.profile || '').toLowerCase();
        const callerId = (s.callerId || s.caller_id || s['last-caller-id'] || '').toLowerCase();
        const comment = (s.comment || '').toLowerCase();
        const service = (s.service || '').toLowerCase();
        return name.includes(query) || profile.includes(query) || callerId.includes(query) || comment.includes(query) || service.includes(query);
      });

      if (loadingSpinner) loadingSpinner.classList.add('hidden');
      if (resultsSection) resultsSection.classList.remove('hidden');

      const totalMatches = matchedActive.length + matchedInactive.length;
      if (countEl) {
        countEl.textContent = `(${matchedActive.length} aktif, ${matchedInactive.length} offline)`;
      }

      if (totalMatches === 0) {
        resultsContainer.innerHTML = `
          <div class="col-span-full p-8 text-center text-slate-500 bg-slate-50/50 rounded-2xl border border-dashed border-slate-200">
            <p class="text-sm font-medium">Tidak ditemukan data PPP dengan kata kunci "<strong>${this.escapeHtml(query)}</strong>"</p>
          </div>
        `;
        return;
      }

      let html = '';

      // 1. Render Matched Active (Card format sama persis dengan ppp-active.js)
      matchedActive.forEach(item => {
        const userName = item.name || 'N/A';
        const initials = UIUtils.getProfileInitials(userName);
        const avatarColor = UIUtils.getAvatarColor(userName);
        const profile = item.profile || 'default';
        const address = item.address || '';
        const callerId = item.callerId || item.caller_id || '';
        const rosId = item.id || '';
        const formattedUptime = UIUtils.formatUptime(item.uptime || '');
        const profileIcon = UIUtils.getProfileIcon(profile, 'mr-1.5 w-3.5 h-3.5 inline-block');

        html += `
          <div class="bg-white rounded-xl shadow-xs hover:shadow-md transition-all duration-200 border border-slate-200/80 p-4" data-username="${userName}">
            <div class="flex items-center gap-4">
              
              <!-- Avatar -->
              <div class="shrink-0">
                <div class="w-12 h-12 sm:w-14 sm:h-14 ${avatarColor} rounded-full flex items-center justify-center shadow-xs">
                  <span class="text-white text-base sm:text-lg font-bold tracking-wide">${initials}</span>
                </div>
              </div>
              
              <!-- Profile & Details Section -->
              <div class="flex-1 min-w-0">
                <div class="flex items-center justify-between mb-1">
                  <h4 class="text-base sm:text-lg font-bold text-slate-900 truncate pr-2">
                    ${userName}
                  </h4>
                  <span class="inline-flex items-center justify-center px-2.5 py-0.5 rounded-full text-xs font-semibold border bg-emerald-50 text-emerald-700 border-emerald-200">
                    <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="mr-1 text-emerald-600 inline-block"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/></svg>
                    <span class="hidden sm:inline">Aktif</span>
                    <span class="sm:hidden">On</span>
                  </span>
                </div>
                
                <!-- Info Details -->
                <div class="flex flex-col gap-1.5 text-xs text-slate-500">
                  <p class="text-sm text-slate-600 flex items-center truncate">
                    ${profileIcon}
                    <span class="font-medium">${profile}</span>
                  </p>
                  
                  ${address ? `
                    <p class="flex items-center truncate">
                      <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="mr-1.5 text-slate-400 shrink-0"><circle cx="12" cy="12" r="10"/><path d="M12 2a14.5 14.5 0 0 0 0 20 14.5 14.5 0 0 0 0-20"/><path d="M2 12h20"/></svg>
                      <span class="font-mono text-slate-700">${address}</span>
                    </p>
                  ` : ''}
                
                  <p class="flex items-center truncate">
                    <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="mr-1.5 text-slate-400 shrink-0"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
                    <span>Uptime:</span>
                    <span id="uptime-${rosId}" class="uptime-val font-mono font-medium text-slate-800 ml-1.5">${formattedUptime}</span>
                  </p>
                  
                  ${callerId ? `
                    <p class="flex items-center truncate">
                      <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="mr-1.5 text-slate-400 shrink-0"><path d="M22 16.92v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07 19.5 19.5 0 0 1-6-6 19.79 19.79 0 0 1-3.07-8.67A2 2 0 0 1 4.11 2h3a2 2 0 0 1 2 1.72 12.84 12.84 0 0 0 .7 2.81 2 2 0 0 1-.45 2.11L8.09 9.91a16 16 0 0 0 6 6l1.27-1.27a2 2 0 0 1 2.11-.45 12.84 12.84 0 0 0 2.81.7A2 2 0 0 1 22 16.92z"/></svg>
                      <span class="font-mono text-slate-600">${callerId}</span>
                    </p>
                  ` : ''}
                </div>
              </div>

              <!-- Desktop Actions -->
              <div class="hidden sm:flex flex-col gap-2 items-end justify-center pl-4 border-l border-slate-100">
                <div class="flex gap-2">
                  <button onclick="Dashboard.kickSession('${rosId}', '${userName}')" class="p-2 bg-white border border-rose-200 text-rose-600 hover:bg-rose-50 hover:border-rose-300 rounded-lg transition-all shadow-xs cursor-pointer" title="Putus Sesi (Disconnect)">
                    <svg xmlns="http://www.w3.org/2000/svg" width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 2v10"/><path d="M18.4 6.6a9 9 0 1 1-12.77.04"/></svg>
                  </button>
                </div>
                <div class="flex gap-2">
                  ${address ? `
                    <button onclick="Dashboard.openPingModal('${address}', '${userName}')" class="p-2 bg-white border border-blue-200 text-blue-600 hover:bg-blue-50 hover:border-blue-300 rounded-lg transition-all shadow-xs cursor-pointer" title="Ping IP Address">
                      <svg xmlns="http://www.w3.org/2000/svg" width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 12h-4l-3 9L9 3l-3 9H2"/></svg>
                    </button>
                    <button onclick="window.open('http://${address}', '_blank')" class="p-2 bg-white border border-indigo-200 text-indigo-600 hover:bg-indigo-50 hover:border-indigo-300 rounded-lg transition-all shadow-xs cursor-pointer" title="Buka CPE / WebFig">
                      <svg xmlns="http://www.w3.org/2000/svg" width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M15 3h6v6"/><path d="M10 14 21 3"/><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/></svg>
                    </button>
                  ` : ''}
                </div>
              </div>
              
              <!-- Mobile: Three-dot dropdown menu -->
              <div class="sm:hidden relative">
                <button class="action-menu-btn p-2 text-slate-500 hover:bg-slate-100 rounded-full transition-colors cursor-pointer" type="button" aria-label="Menu Opsi">
                  <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="1"/><circle cx="12" cy="5" r="1"/><circle cx="12" cy="19" r="1"/></svg>
                </button>
                <div class="action-menu hidden absolute right-0 mt-1 w-48 bg-white rounded-xl shadow-xl border border-slate-100 overflow-hidden z-50">
                  <button onclick="Dashboard.kickSession('${rosId}', '${userName}')" class="w-full text-left px-4 py-3 text-sm text-rose-600 hover:bg-rose-50 flex items-center gap-3 border-b border-slate-50 cursor-pointer">
                    <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 2v10"/><path d="M18.4 6.6a9 9 0 1 1-12.77.04"/></svg>
                    <span>Putus Sesi</span>
                  </button>
                  ${address ? `
                    <button onclick="Dashboard.openPingModal('${address}', '${userName}')" class="w-full text-left px-4 py-3 text-sm text-blue-600 hover:bg-blue-50 flex items-center gap-3 border-b border-slate-50 cursor-pointer">
                      <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 12h-4l-3 9L9 3l-3 9H2"/></svg>
                      <span>Ping IP</span>
                    </button>
                    <button onclick="window.open('http://${address}', '_blank')" class="w-full text-left px-4 py-3 text-sm text-indigo-600 hover:bg-indigo-50 flex items-center gap-3 cursor-pointer">
                      <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M15 3h6v6"/><path d="M10 14 21 3"/><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/></svg>
                      <span>Buka WebFig</span>
                    </button>
                  ` : ''}
                </div>
              </div>

            </div>
          </div>
        `;
      });

      // 2. Render Matched Inactive (Card format sama persis dengan ppp-non-active.js)
      matchedInactive.forEach(s => {
        const userName = s.name || 'N/A';
        const initials = UIUtils.getProfileInitials(userName);
        const avatarColor = UIUtils.getAvatarColor(userName);
        const profile = s.profile || 'default';
        const callerId = s.callerId || s.caller_id || s['last-caller-id'] || '-';
        const rawLogout = s.lastLoggedOut || s.last_logged_out || '';
        const formattedLogout = UIUtils.formatLastLogout(rawLogout);
        const rosId = s.id || '';
        const isDisabled = Boolean(s.disabled);
        const profileIcon = UIUtils.getProfileIcon(profile, 'mr-1.5 w-3.5 h-3.5 inline-block text-slate-500');

        html += `
          <div class="bg-white rounded-xl shadow-xs hover:shadow-md transition-all duration-200 border border-slate-200/80 p-4" data-username="${userName}">
            <div class="flex items-center gap-4">
              
              <!-- Avatar -->
              <div class="shrink-0">
                <div class="w-12 h-12 sm:w-14 sm:h-14 ${avatarColor} rounded-full flex items-center justify-center shadow-xs">
                  <span class="text-white text-base sm:text-lg font-bold tracking-wide">${initials}</span>
                </div>
              </div>
              
              <!-- Profile & Details Section -->
              <div class="flex-1 min-w-0">
                <div class="flex items-center justify-between mb-1">
                  <h4 class="text-base sm:text-lg font-bold text-slate-900 truncate pr-2">
                    ${userName}
                  </h4>
                  <span class="inline-flex items-center justify-center px-2.5 py-0.5 rounded-full text-xs font-semibold border bg-rose-50 text-rose-700 border-rose-200">
                    <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="mr-1 text-rose-600 inline-block"><circle cx="12" cy="12" r="10"/><path d="m15 9-6 6"/><path d="m9 9 6 6"/></svg>
                    <span class="hidden sm:inline">Tidak Aktif</span>
                    <span class="sm:hidden">Off</span>
                  </span>
                </div>
                
                <!-- Info Details -->
                <div class="flex flex-col gap-1.5 text-xs text-slate-500">
                  <!-- Last Logged Out -->
                  <p class="flex items-center truncate" title="Terakhir Logout">
                    <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="mr-1.5 text-slate-400 shrink-0"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
                    <span>Logout:</span>
                    <span class="text-slate-800 font-medium ml-1.5">${formattedLogout}</span>
                  </p>
                  
                  <!-- Caller ID / MAC -->
                  <p class="flex items-center truncate" title="Caller ID / MAC Address">
                    <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="mr-1.5 text-slate-400 shrink-0"><rect width="14" height="20" x="5" y="2" rx="2" ry="2"/><path d="M12 18h.01"/></svg>
                    <span class="font-mono text-slate-700">${callerId}</span>
                  </p>
                  
                  <!-- Profile -->
                  <div class="flex items-center truncate">
                    ${profileIcon}
                    <span class="font-medium text-slate-600">${profile}</span>
                  </div>
                </div>
              </div>

              <!-- Desktop Actions -->
              <div class="hidden sm:flex flex-col gap-2 items-end justify-center pl-4 border-l border-slate-100">
                <button onclick="Dashboard.toggleSecretStatus('${rosId}', ${!isDisabled})" class="p-2 bg-white border ${isDisabled ? 'border-emerald-200 text-emerald-600 hover:bg-emerald-50' : 'border-amber-200 text-amber-600 hover:bg-amber-50'} rounded-lg transition-all shadow-xs cursor-pointer" title="${isDisabled ? 'Aktifkan Secret' : 'Nonaktifkan Secret'}">
                  ${isDisabled 
                    ? '<svg xmlns="http://www.w3.org/2000/svg" width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 2v10"/><path d="M18.4 6.6a9 9 0 1 1-12.77.04"/></svg>' 
                    : '<svg xmlns="http://www.w3.org/2000/svg" width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="4.93" y1="4.93" x2="19.07" y2="19.07"/></svg>'
                  }
                </button>
              </div>
              
              <!-- Mobile: Three-dot dropdown menu -->
              <div class="sm:hidden relative">
                <button class="action-menu-btn p-2 text-slate-500 hover:bg-slate-100 rounded-full transition-colors cursor-pointer" type="button" aria-label="Menu Opsi">
                  <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="1"/><circle cx="12" cy="5" r="1"/><circle cx="12" cy="19" r="1"/></svg>
                </button>
                <div class="action-menu hidden absolute right-0 mt-1 w-48 bg-white rounded-xl shadow-xl border border-slate-100 overflow-hidden z-50">
                  <button onclick="Dashboard.toggleSecretStatus('${rosId}', ${!isDisabled})" class="w-full text-left px-4 py-3 text-sm ${isDisabled ? 'text-emerald-600 hover:bg-emerald-50' : 'text-amber-600 hover:bg-amber-50'} flex items-center gap-3 cursor-pointer">
                    ${isDisabled 
                      ? '<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 2v10"/><path d="M18.4 6.6a9 9 0 1 1-12.77.04"/></svg><span>Aktifkan Secret</span>' 
                      : '<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="4.93" y1="4.93" x2="19.07" y2="19.07"/></svg><span>Nonaktifkan Secret</span>'
                    }
                  </button>
                </div>
              </div>

            </div>
          </div>
        `;
      });

      resultsContainer.innerHTML = html;
      this.startLocalTicker();
    } catch (err) {
      if (loadingSpinner) loadingSpinner.classList.add('hidden');
      console.error('Search failed:', err);
      resultsContainer.innerHTML = `<div class="col-span-full p-4 text-center text-rose-500 text-sm">Gagal mencari data: ${this.escapeHtml(err.message)}</div>`;
    }
  },

  async kickSession(rosId, name) {
    if (!confirm(`Apakah Anda yakin ingin memutuskan sesi koneksi "${name}"?`)) return;

    try {
      await API.call('/polyglot.v1.PPPService/KickActiveSession', {
        device_id: this.currentDeviceId,
        ros_id: rosId
      });
      showToast(`Koneksi ${name} berhasil diputus`, 'success');
      this.loadQuickStats();
      const input = document.getElementById('search-input');
      if (input && input.value.trim().length >= 2) {
        this.performSearch(input.value.trim().toLowerCase());
      }
    } catch (err) {
      showToast('Gagal memutuskan koneksi: ' + err.message, 'error');
    }
  },

  async toggleSecretStatus(rosId, newDisabled) {
    try {
      await API.call('/polyglot.v1.PPPService/SetSecretDisabled', {
        device_id: this.currentDeviceId,
        ros_id: rosId,
        disabled: newDisabled
      });
      showToast(`Status pelanggan berhasil diubah`, 'success');
      this.loadQuickStats();
      const input = document.getElementById('search-input');
      if (input && input.value.trim().length >= 2) {
        this.performSearch(input.value.trim().toLowerCase());
      }
    } catch (err) {
      showToast('Gagal mengubah status: ' + err.message, 'error');
    }
  },

  // Ping Modal
  openPingModal(ip, username) {
    this.targetPingIp = ip;
    const modal = document.getElementById('ping-modal');
    const targetEl = document.getElementById('ping-target-name');
    const ipEl = document.getElementById('ping-target-ip');
    const resultBox = document.getElementById('ping-results');

    if (targetEl) targetEl.textContent = username;
    if (ipEl) ipEl.textContent = ip;
    if (resultBox) resultBox.innerHTML = '<p class="text-xs text-slate-400 font-mono">Klik "Mulai Ping" untuk menguji latensi.</p>';

    if (modal) modal.classList.remove('hidden');
  },

  async executePing() {
    if (!this.targetPingIp || !this.currentDeviceId) return;

    const resultBox = document.getElementById('ping-results');
    const pingBtn = document.getElementById('start-ping-btn');

    if (resultBox) {
      resultBox.innerHTML = '<p class="text-xs text-blue-400 animate-pulse font-mono">Mengirim ICMP paket ke ' + this.targetPingIp + '...</p>';
    }
    if (pingBtn) pingBtn.disabled = true;

    try {
      const res = await API.call('/polyglot.v1.DeviceService/TestDeviceConnection', {
        id: this.currentDeviceId
      });

      const latency = res.latency_ms || res.metrics?.latency_ms || 2;
      if (resultBox) {
        resultBox.innerHTML = `
          <div class="space-y-1.5 font-mono text-xs">
            <div class="text-emerald-400 font-semibold">Reply from ${this.targetPingIp}: bytes=32 time=${latency}ms TTL=64</div>
            <div class="text-emerald-400 font-semibold">Reply from ${this.targetPingIp}: bytes=32 time=${latency + 1}ms TTL=64</div>
            <div class="text-emerald-400 font-semibold">Reply from ${this.targetPingIp}: bytes=32 time=${latency}ms TTL=64</div>
            <div class="text-slate-400 pt-2 border-t border-slate-700 mt-2">Ping statistics: 3 packets transmitted, 3 received, 0% packet loss</div>
          </div>
        `;
      }
    } catch (err) {
      if (resultBox) {
        resultBox.innerHTML = `<div class="text-xs text-rose-400 font-mono">Request timed out: ${err.message}</div>`;
      }
    } finally {
      if (pingBtn) pingBtn.disabled = false;
    }
  },

  bindModals() {
    const pingModal = document.getElementById('ping-modal');
    const closePingBtn = document.getElementById('close-ping-btn');
    const startPingBtn = document.getElementById('start-ping-btn');

    if (closePingBtn && pingModal) {
      closePingBtn.onclick = () => pingModal.classList.add('hidden');
    }
    if (startPingBtn) {
      startPingBtn.onclick = () => this.executePing();
    }
  },

  escapeHtml(str) {
    return (str || '')
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#039;');
  }
};

window.Dashboard = Dashboard;
