/**
 * Polyglot Web Dial Dashboard Controller
 * Streaming real-time system resources (CPU, RAM, HDD, Uptime, OS) via StreamSystemSnapshot
 * and managing quick PPP statistics and search.
 */

const Dashboard = {
  currentDeviceId: null,
  resourceStreamController: null,

  init() {
    RouterSelector.onDeviceChange((deviceId) => {
      if (!deviceId) return;
      this.currentDeviceId = deviceId;
      this.startResourceStream();
      this.loadInterfaceList();
      this.loadQuickStats();
    });

    this.bindSearch();

    window.addEventListener('beforeunload', () => {
      this.cleanupStreams();
    });
  },

  cleanupStreams() {
    if (this.resourceStreamController) {
      this.resourceStreamController.abort();
      this.resourceStreamController = null;
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
   * Mengambil snapshot interface list untuk menampilkan status link
   */
  async loadInterfaceList() {
    if (!this.currentDeviceId) return;
    const ifacesContainer = document.getElementById('sys-interfaces');
    if (!ifacesContainer) return;

    try {
      const res = await API.call('/polyglot.v1.DeviceService/TestDeviceConnection', {
        id: this.currentDeviceId
      });

      const list = res.interface_list || res.metrics?.interface_list || [];
      if (list.length > 0) {
        ifacesContainer.innerHTML = list.map(ifc => {
          const run = Boolean(ifc.running);
          const dis = Boolean(ifc.disabled);
          return `
            <span class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg text-xs font-mono font-medium ${
              dis ? 'bg-slate-100 text-slate-400' :
              run ? 'bg-emerald-50 text-emerald-700 border border-emerald-200/80' :
              'bg-slate-100 text-slate-600 border border-slate-200/60'
            }" title="${ifc.name} (${dis ? 'Disabled' : run ? 'Link Up' : 'Link Down'})">
              <span class="w-2 h-2 rounded-full ${dis ? 'bg-slate-300' : run ? 'bg-emerald-500 animate-pulse' : 'bg-slate-400'}"></span>
              ${ifc.name}
            </span>
          `;
        }).join('');
      } else {
        const names = res.interfaces || res.metrics?.interfaces || [];
        if (names.length > 0) {
          ifacesContainer.innerHTML = names.map(n => `
            <span class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg text-xs font-mono bg-slate-100 text-slate-700 border border-slate-200">
              <span class="w-2 h-2 rounded-full bg-emerald-500"></span>
              ${n}
            </span>
          `).join('');
        }
      }
    } catch (err) {
      console.warn('Could not load interface list:', err);
    }
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
        return;
      }
      timeout = setTimeout(() => this.performSearch(query), 300);
    });
  },

  async performSearch(query) {
    if (!this.currentDeviceId) return;
    const resultsContainer = document.getElementById('search-results');
    const resultsSection = document.getElementById('search-results-section');
    const loadingSpinner = document.getElementById('search-loading');

    if (loadingSpinner) loadingSpinner.classList.remove('hidden');

    try {
      const [activeRes, secretsRes] = await Promise.all([
        API.call('/polyglot.v1.PPPService/ListActiveSessions', { device_id: this.currentDeviceId }),
        API.call('/polyglot.v1.PPPService/ListSecrets', { device_id: this.currentDeviceId })
      ]);

      const activeList = activeRes.sessions || [];
      const secretList = secretsRes.secrets || [];

      const matchedActive = activeList.filter(s => 
        (s.name && s.name.toLowerCase().includes(query)) ||
        (s.caller_id && s.caller_id.toLowerCase().includes(query)) ||
        (s.address && s.address.toLowerCase().includes(query))
      );

      const matchedSecrets = secretList.filter(s =>
        (s.name && s.name.toLowerCase().includes(query)) ||
        (s.comment && s.comment.toLowerCase().includes(query)) ||
        (s.remote_address && s.remote_address.toLowerCase().includes(query))
      );

      if (loadingSpinner) loadingSpinner.classList.add('hidden');
      if (resultsSection) resultsSection.classList.remove('hidden');

      if (matchedActive.length === 0 && matchedSecrets.length === 0) {
        resultsContainer.innerHTML = `<div class="p-8 text-center text-slate-500"><p class="text-sm">Tidak ditemukan data PPP dengan kata kunci "<strong>${query}</strong>"</p></div>`;
        return;
      }

      let html = '<div class="grid grid-cols-1 md:grid-cols-2 gap-4">';
      matchedActive.forEach(item => {
        html += `
          <div class="bg-emerald-50/40 border border-emerald-200 rounded-xl p-4 flex flex-col justify-between shadow-xs">
            <div>
              <div class="flex items-center justify-between mb-2">
                <span class="font-bold text-slate-900 text-base">${item.name}</span>
                <span class="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-emerald-100 text-emerald-800">
                  <span class="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse"></span> Online
                </span>
              </div>
              <div class="grid grid-cols-2 gap-2 text-xs text-slate-600">
                <div><span class="text-slate-400">IP:</span> ${item.address || '-'}</div>
                <div><span class="text-slate-400">Caller ID:</span> ${item.caller_id || '-'}</div>
                <div><span class="text-slate-400">Service:</span> ${item.service || 'pppoe'}</div>
                <div><span class="text-slate-400">Uptime:</span> ${item.uptime || '-'}</div>
              </div>
            </div>
            <div class="mt-3 pt-3 border-t border-emerald-100 flex items-center justify-between">
              <a href="ppp-active.html?search=${encodeURIComponent(item.name)}" class="text-xs font-semibold text-emerald-700 hover:underline flex items-center gap-1">Buka di PPP Aktif &rarr;</a>
              ${item.address ? `<button onclick="window.open('http://${item.address}', '_blank')" class="text-xs bg-white border border-emerald-300 text-emerald-700 px-2.5 py-1 rounded-md hover:bg-emerald-50 font-medium">Buka CPE</button>` : ''}
            </div>
          </div>
        `;
      });

      matchedSecrets.forEach(item => {
        if (activeList.some(a => a.name === item.name)) return;
        html += `
          <div class="bg-slate-50 border border-slate-200 rounded-xl p-4 flex flex-col justify-between shadow-xs">
            <div>
              <div class="flex items-center justify-between mb-2">
                <span class="font-bold text-slate-900 text-base">${item.name}</span>
                <span class="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-slate-200 text-slate-700">Offline</span>
              </div>
              <div class="grid grid-cols-2 gap-2 text-xs text-slate-600">
                <div><span class="text-slate-400">Profile:</span> ${item.profile || '-'}</div>
                <div><span class="text-slate-400">Remote IP:</span> ${item.remote_address || '-'}</div>
                <div class="col-span-2"><span class="text-slate-400">Comment:</span> ${item.comment || '-'}</div>
              </div>
            </div>
            <div class="mt-3 pt-3 border-t border-slate-200 flex items-center justify-between">
              <a href="ppp-secrets.html?search=${encodeURIComponent(item.name)}" class="text-xs font-semibold text-blue-600 hover:underline">Buka di Secrets &rarr;</a>
            </div>
          </div>
        `;
      });

      html += '</div>';
      resultsContainer.innerHTML = html;
    } catch (err) {
      if (loadingSpinner) loadingSpinner.classList.add('hidden');
      console.error('Search failed:', err);
      resultsContainer.innerHTML = `<div class="p-4 text-center text-rose-500 text-sm">Gagal mencari data: ${err.message}</div>`;
    }
  }
};

window.Dashboard = Dashboard;
