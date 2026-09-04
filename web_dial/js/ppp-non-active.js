/**
 * Polyglot Web Dial PPP Non-Active (Offline) Controller
 */

const PPPNonActive = {
  currentDeviceId: null,
  allSecrets: [],
  filteredSecrets: [],
  currentPage: 1,
  pageSize: 25,
  streamAbortController: null,

  init() {
    RouterSelector.onDeviceChange((deviceId) => {
      if (!deviceId) return;
      this.currentDeviceId = deviceId;
      this.allSecrets = [];
      this.cleanupStream();
      this.loadInactive();
    });

    this.bindControls();
  },

  cleanupStream() {
    if (this.streamAbortController) {
      try {
        this.streamAbortController.abort();
      } catch (_) {}
      this.streamAbortController = null;
    }
  },

  async loadInactive() {
    const loadingEl = document.getElementById('loading-state');
    const resultsGrid = document.getElementById('results-grid');
    const emptyEl = document.getElementById('empty-state');
    const paginationEl = document.getElementById('pagination-container');

    this.cleanupStream();

    if (this.allSecrets.length === 0) {
      if (loadingEl) loadingEl.classList.remove('hidden');
      if (resultsGrid) resultsGrid.classList.add('hidden');
      if (emptyEl) emptyEl.classList.add('hidden');
      if (paginationEl) paginationEl.classList.add('hidden');
    }

    this.streamAbortController = new AbortController();
    const signal = this.streamAbortController.signal;

    try {
      await API.stream(
        '/polyglot.v1.PPPService/StreamInactiveSecrets',
        {
          device_id: this.currentDeviceId
        },
        (frame) => {
          if (!frame) return;
          if (loadingEl) loadingEl.classList.add('hidden');
          this.allSecrets = frame.secrets || [];
          this.applyFilter();
        },
        (err) => {
          if (err && err.name === 'AbortError') return;
          console.warn('Inactive secrets stream closed or failed:', err);
          if (loadingEl) loadingEl.classList.add('hidden');
        },
        signal
      );
    } catch (err) {
      if (err && err.name === 'AbortError') return;
      console.error('Failed to start inactive secrets stream:', err);
      if (loadingEl) loadingEl.classList.add('hidden');
      // Fallback to one-shot fetch if streaming fails
      try {
        const res = await API.call('/polyglot.v1.PPPService/ListInactiveSecrets', {
          device_id: this.currentDeviceId
        });
        this.allSecrets = res.secrets || [];
        this.applyFilter();
      } catch (fallbackErr) {
        showToast('Gagal memuat pelanggan offline: ' + fallbackErr.message, 'error');
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
      refreshBtn.onclick = () => this.loadInactive();
    }
  },

  applyFilter() {
    const query = (document.getElementById('search-input')?.value || '').toLowerCase().trim();

    this.filteredSecrets = this.allSecrets.filter((s) => {
      if (!query) return true;
      const name = (s.name || '').toLowerCase();
      const profile = (s.profile || '').toLowerCase();
      const callerId = (s.callerId || s.caller_id || s['last-caller-id'] || '').toLowerCase();

      return name.includes(query) || profile.includes(query) || callerId.includes(query);
    });

    const countEl = document.getElementById('results-count');
    if (countEl) countEl.textContent = `(${this.filteredSecrets.length} offline)`;

    this.render();
  },

  getCurrentPageItems() {
    const startIdx = (this.currentPage - 1) * this.pageSize;
    return this.filteredSecrets.slice(startIdx, startIdx + this.pageSize);
  },

  render() {
    const resultsGrid = document.getElementById('results-grid');
    const emptyEl = document.getElementById('empty-state');
    const paginationEl = document.getElementById('pagination-container');

    if (!resultsGrid) return;

    if (this.filteredSecrets.length === 0) {
      resultsGrid.classList.add('hidden');
      if (emptyEl) emptyEl.classList.remove('hidden');
      if (paginationEl) paginationEl.classList.add('hidden');
      return;
    }

    resultsGrid.classList.remove('hidden');
    if (emptyEl) emptyEl.classList.add('hidden');

    const totalPages = Math.ceil(this.filteredSecrets.length / this.pageSize);
    const pageItems = this.getCurrentPageItems();

    resultsGrid.innerHTML = pageItems.map((s) => {
      const userName = s.name || 'N/A';
      const initials = UIUtils.getProfileInitials(userName);
      const avatarColor = UIUtils.getAvatarColor(userName);
      const profile = s.profile || 'default';
      const callerId = s.callerId || s.caller_id || s['last-caller-id'] || '-';
      const rawLogout = s.lastLoggedOut || s.last_logged_out || '';
      const formattedLogout = UIUtils.formatLastLogout(rawLogout);
      const rosId = s.id || '';
      const isDisabled = s.disabled || false;

      const profileIcon = UIUtils.getProfileIcon(profile, 'mr-1.5 w-3.5 h-3.5 inline-block text-slate-500');

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
              <button onclick="PPPNonActive.toggleSecretStatus('${rosId}', ${!isDisabled})" class="p-2 bg-white border ${isDisabled ? 'border-emerald-200 text-emerald-600 hover:bg-emerald-50' : 'border-amber-200 text-amber-600 hover:bg-amber-50'} rounded-lg transition-all shadow-xs cursor-pointer" title="${isDisabled ? 'Aktifkan Secret' : 'Nonaktifkan Secret'}">
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
                <button onclick="PPPNonActive.toggleSecretStatus('${rosId}', ${!isDisabled})" class="w-full text-left px-4 py-3 text-sm ${isDisabled ? 'text-emerald-600 hover:bg-emerald-50' : 'text-amber-600 hover:bg-amber-50'} flex items-center gap-3 cursor-pointer">
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
    }).join('');

    this.renderPagination(totalPages);
  },

  async toggleSecretStatus(rosId, newDisabled) {
    try {
      await API.call('/polyglot.v1.PPPService/SetSecretDisabled', {
        device_id: this.currentDeviceId,
        ros_id: rosId,
        disabled: newDisabled
      });
      showToast(`Status pelanggan berhasil diubah`, 'success');
      this.loadInactive();
    } catch (err) {
      showToast('Gagal mengubah status: ' + err.message, 'error');
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
    const total = this.filteredSecrets.length;
    const startIdx = (this.currentPage - 1) * this.pageSize + 1;
    const endIdx = Math.min(this.currentPage * this.pageSize, total);

    el.innerHTML = `
      <div class="flex flex-col sm:flex-row justify-between items-center gap-4 text-xs text-slate-500 border-t border-slate-200/80 pt-4">
        <p>Menampilkan <span class="font-semibold text-slate-700">${startIdx}</span> - <span class="font-semibold text-slate-700">${endIdx}</span> dari <span class="font-semibold text-slate-700">${total}</span> pelanggan offline</p>
        <div class="flex items-center gap-1.5">
          <button ${this.currentPage <= 1 ? 'disabled' : ''} onclick="PPPNonActive.goToPage(${this.currentPage - 1})" class="px-3 py-1.5 rounded-lg border border-slate-300 bg-white hover:bg-slate-50 text-slate-700 disabled:opacity-40 disabled:cursor-not-allowed shadow-xs cursor-pointer">Sebelumnya</button>
          <span class="px-2 text-slate-700 font-semibold">${this.currentPage} / ${totalPages}</span>
          <button ${this.currentPage >= totalPages ? 'disabled' : ''} onclick="PPPNonActive.goToPage(${this.currentPage + 1})" class="px-3 py-1.5 rounded-lg border border-slate-300 bg-white hover:bg-slate-50 text-slate-700 disabled:opacity-40 disabled:cursor-not-allowed shadow-xs cursor-pointer">Berikutnya</button>
        </div>
      </div>
    `;
  },

  goToPage(page) {
    this.currentPage = page;
    this.render();
  }
};

window.PPPNonActive = PPPNonActive;
