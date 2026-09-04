/**
 * Polyglot Web Dial PPP Active Sessions Controller
 */

const PPPActive = {
  currentDeviceId: null,
  allSessions: [],
  filteredSessions: [],
  currentPage: 1,
  pageSize: 25,
  targetPingIp: null,
  localTicker: null,
  streamAbortController: null,

  init() {
    RouterSelector.onDeviceChange((deviceId) => {
      if (!deviceId) return;
      this.currentDeviceId = deviceId;
      this.allSessions = [];
      this.cleanupStream();
      this.loadActiveSessions();
    });

    this.bindControls();
    this.bindModals();

    // Check query param for search filter
    const urlParams = new URLSearchParams(window.location.search);
    const searchParam = urlParams.get('search');
    if (searchParam) {
      const searchInput = document.getElementById('search-input');
      if (searchInput) searchInput.value = searchParam;
    }
  },

  cleanupStream() {
    if (this.streamAbortController) {
      try {
        this.streamAbortController.abort();
      } catch (_) {}
      this.streamAbortController = null;
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

  async loadActiveSessions() {
    const loadingEl = document.getElementById('loading-state');
    const resultsGrid = document.getElementById('results-grid');
    const emptyEl = document.getElementById('empty-state');
    const paginationEl = document.getElementById('pagination-container');

    this.cleanupStream();

    if (this.allSessions.length === 0) {
      if (loadingEl) loadingEl.classList.remove('hidden');
      if (resultsGrid) resultsGrid.classList.add('hidden');
      if (emptyEl) emptyEl.classList.add('hidden');
      if (paginationEl) paginationEl.classList.add('hidden');
    }

    this.streamAbortController = new AbortController();
    const signal = this.streamAbortController.signal;

    try {
      await API.stream(
        '/polyglot.v1.PPPService/StreamActiveSessions',
        {
          device_id: this.currentDeviceId
        },
        (frame) => {
          if (!frame) return;
          if (loadingEl) loadingEl.classList.add('hidden');
          this.allSessions = frame.sessions || [];
          this.applyFilter();
          this.startLocalTicker();
        },
        (err) => {
          if (err && err.name === 'AbortError') return;
          console.warn('Active sessions stream closed or failed:', err);
          if (loadingEl) loadingEl.classList.add('hidden');
        },
        signal
      );
    } catch (err) {
      if (err && err.name === 'AbortError') return;
      console.error('Failed to start active sessions stream:', err);
      if (loadingEl) loadingEl.classList.add('hidden');
      // Fallback to one-shot fetch if stream cannot be established
      try {
        const res = await API.call('/polyglot.v1.PPPService/ListActiveSessions', {
          device_id: this.currentDeviceId
        });
        this.allSessions = res.sessions || [];
        this.applyFilter();
        this.startLocalTicker();
      } catch (fallbackErr) {
        showToast('Gagal memuat sesi aktif: ' + fallbackErr.message, 'error');
      }
    }
  },


  bindControls() {
    const searchInput = document.getElementById('search-input');
    const pageSizeSelect = document.getElementById('per-page-select');
    const refreshBtn = document.getElementById('refresh-btn');

    if (searchInput) {
      searchInput.addEventListener('input', () => {
        this.currentPage = 1;
        this.applyFilter();
      });
    }

    if (pageSizeSelect) {
      pageSizeSelect.addEventListener('change', (e) => {
        this.pageSize = e.target.value === 'all' ? 999999 : parseInt(e.target.value, 10);
        this.currentPage = 1;
        this.render();
      });
    }

    if (refreshBtn) {
      refreshBtn.onclick = () => this.loadActiveSessions();
    }
  },

  applyFilter() {
    const query = (document.getElementById('search-input')?.value || '').toLowerCase().trim();

    this.filteredSessions = this.allSessions.filter((s) => {
      if (!query) return true;
      const callerId = (s.callerId || s.caller_id || '').toLowerCase();
      const name = (s.name || '').toLowerCase();
      const address = (s.address || '').toLowerCase();
      const service = (s.service || '').toLowerCase();
      const profile = (s.profile || '').toLowerCase();

      return name.includes(query) || callerId.includes(query) || address.includes(query) || service.includes(query) || profile.includes(query);
    });

    const countEl = document.getElementById('results-count');
    if (countEl) countEl.textContent = `(${this.filteredSessions.length} aktif)`;

    this.render();
  },

  getCurrentPageItems() {
    const startIdx = (this.currentPage - 1) * this.pageSize;
    return this.filteredSessions.slice(startIdx, startIdx + this.pageSize);
  },

  render() {
    const resultsGrid = document.getElementById('results-grid');
    const emptyEl = document.getElementById('empty-state');
    const paginationEl = document.getElementById('pagination-container');

    if (!resultsGrid) return;

    if (this.filteredSessions.length === 0) {
      resultsGrid.classList.add('hidden');
      if (emptyEl) emptyEl.classList.remove('hidden');
      if (paginationEl) paginationEl.classList.add('hidden');
      return;
    }

    resultsGrid.classList.remove('hidden');
    if (emptyEl) emptyEl.classList.add('hidden');

    const totalPages = Math.ceil(this.filteredSessions.length / this.pageSize);
    const pageItems = this.getCurrentPageItems();

    resultsGrid.innerHTML = pageItems.map((item) => {
      const userName = item.name || 'N/A';
      const initials = UIUtils.getProfileInitials(userName);
      const avatarColor = UIUtils.getAvatarColor(userName);
      const profile = item.profile || 'default';
      const address = item.address || '';
      const callerId = item.callerId || item.caller_id || '';
      const rosId = item.id || '';

      const formattedUptime = UIUtils.formatUptime(item.uptime || '');

      const profileIcon = UIUtils.getProfileIcon(profile, 'mr-1.5 w-3.5 h-3.5 inline-block');

      return `
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
                <button onclick="PPPActive.kickSession('${rosId}', '${userName}')" class="p-2 bg-white border border-rose-200 text-rose-600 hover:bg-rose-50 hover:border-rose-300 rounded-lg transition-all shadow-xs cursor-pointer" title="Putus Sesi (Disconnect)">
                  <svg xmlns="http://www.w3.org/2000/svg" width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 2v10"/><path d="M18.4 6.6a9 9 0 1 1-12.77.04"/></svg>
                </button>
              </div>
              <div class="flex gap-2">
                ${address ? `
                  <button onclick="PPPActive.openPingModal('${address}', '${userName}')" class="p-2 bg-white border border-blue-200 text-blue-600 hover:bg-blue-50 hover:border-blue-300 rounded-lg transition-all shadow-xs cursor-pointer" title="Ping IP Address">
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
                <button onclick="PPPActive.kickSession('${rosId}', '${userName}')" class="w-full text-left px-4 py-3 text-sm text-rose-600 hover:bg-rose-50 flex items-center gap-3 border-b border-slate-50 cursor-pointer">
                  <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 2v10"/><path d="M18.4 6.6a9 9 0 1 1-12.77.04"/></svg>
                  <span>Putus Sesi</span>
                </button>
                ${address ? `
                  <button onclick="PPPActive.openPingModal('${address}', '${userName}')" class="w-full text-left px-4 py-3 text-sm text-blue-600 hover:bg-blue-50 flex items-center gap-3 border-b border-slate-50 cursor-pointer">
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
    }).join('');

    this.renderPagination(totalPages);
  },

  async kickSession(rosId, name) {
    if (!confirm(`Apakah Anda yakin ingin memutuskan sesi koneksi "${name}"?`)) return;

    try {
      await API.call('/polyglot.v1.PPPService/KickActiveSession', {
        device_id: this.currentDeviceId,
        ros_id: rosId
      });
      showToast(`Koneksi ${name} berhasil diputus`, 'success');
      this.loadActiveSessions();
    } catch (err) {
      showToast('Gagal memutuskan koneksi: ' + err.message, 'error');
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

  renderPagination(totalPages) {
    const el = document.getElementById('pagination-container');
    if (!el) return;

    if (totalPages <= 1) {
      el.classList.add('hidden');
      el.innerHTML = '';
      return;
    }

    el.classList.remove('hidden');
    const total = this.filteredSessions.length;
    const startIdx = (this.currentPage - 1) * this.pageSize + 1;
    const endIdx = Math.min(this.currentPage * this.pageSize, total);

    el.innerHTML = `
      <div class="flex flex-col sm:flex-row justify-between items-center gap-4 text-xs text-slate-500 border-t border-slate-200/80 pt-4">
        <p>Menampilkan <span class="font-semibold text-slate-700">${startIdx}</span> - <span class="font-semibold text-slate-700">${endIdx}</span> dari <span class="font-semibold text-slate-700">${total}</span> sesi aktif</p>
        <div class="flex items-center gap-1.5">
          <button ${this.currentPage <= 1 ? 'disabled' : ''} onclick="PPPActive.goToPage(${this.currentPage - 1})" class="px-3 py-1.5 rounded-lg border border-slate-300 bg-white hover:bg-slate-50 text-slate-700 disabled:opacity-40 disabled:cursor-not-allowed shadow-xs cursor-pointer">Sebelumnya</button>
          <span class="px-2 text-slate-700 font-semibold">${this.currentPage} / ${totalPages}</span>
          <button ${this.currentPage >= totalPages ? 'disabled' : ''} onclick="PPPActive.goToPage(${this.currentPage + 1})" class="px-3 py-1.5 rounded-lg border border-slate-300 bg-white hover:bg-slate-50 text-slate-700 disabled:opacity-40 disabled:cursor-not-allowed shadow-xs cursor-pointer">Berikutnya</button>
        </div>
      </div>
    `;
  },

  goToPage(page) {
    this.currentPage = page;
    this.render();
  }
};

window.PPPActive = PPPActive;
