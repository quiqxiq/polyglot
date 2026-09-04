/**
 * Polyglot Web Dial PPP Profiles Controller
 */

const PPPProfiles = {
  currentDeviceId: null,
  allProfiles: [],
  filteredProfiles: [],
  currentPage: 1,
  pageSize: 25,

  init() {
    RouterSelector.onDeviceChange((deviceId) => {
      if (!deviceId) return;
      this.currentDeviceId = deviceId;
      this.loadProfiles();
    });

    this.bindControls();
  },

  async loadProfiles() {
    const loadingEl = document.getElementById('loading-state');
    const resultsGrid = document.getElementById('results-grid');
    const emptyEl = document.getElementById('empty-state');
    const paginationEl = document.getElementById('pagination-container');

    if (loadingEl) loadingEl.classList.remove('hidden');
    if (resultsGrid) resultsGrid.classList.add('hidden');
    if (emptyEl) emptyEl.classList.add('hidden');
    if (paginationEl) paginationEl.classList.add('hidden');

    try {
      const res = await API.call('/polyglot.v1.PPPService/ListProfiles', {
        device_id: this.currentDeviceId
      });

      this.allProfiles = res.profiles || [];
      this.applyFilter();
    } catch (err) {
      console.error('Failed to load profiles:', err);
      showToast('Gagal memuat profil: ' + err.message, 'error');
    } finally {
      if (loadingEl) loadingEl.classList.add('hidden');
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
      refreshBtn.onclick = () => this.loadProfiles();
    }
  },

  applyFilter() {
    const query = (document.getElementById('search-input')?.value || '').toLowerCase().trim();

    this.filteredProfiles = this.allProfiles.filter((p) => {
      if (!query) return true;
      const name = (p.name || '').toLowerCase();
      const rateLimit = (p.rateLimit || p.rate_limit || '').toLowerCase();
      const localAddress = (p.localAddress || p.local_address || '').toLowerCase();
      const remoteAddress = (p.remoteAddress || p.remote_address || '').toLowerCase();
      const parentQueue = (p.parentQueue || p.parent_queue || '').toLowerCase();

      return name.includes(query) || rateLimit.includes(query) || localAddress.includes(query) || remoteAddress.includes(query) || parentQueue.includes(query);
    });

    const countEl = document.getElementById('results-count');
    if (countEl) countEl.textContent = `(${this.filteredProfiles.length} profil)`;

    this.render();
  },

  getCurrentPageItems() {
    const startIdx = (this.currentPage - 1) * this.pageSize;
    return this.filteredProfiles.slice(startIdx, startIdx + this.pageSize);
  },

  render() {
    const resultsGrid = document.getElementById('results-grid');
    const emptyEl = document.getElementById('empty-state');
    const paginationEl = document.getElementById('pagination-container');

    if (!resultsGrid) return;

    if (this.filteredProfiles.length === 0) {
      resultsGrid.classList.add('hidden');
      if (emptyEl) emptyEl.classList.remove('hidden');
      if (paginationEl) paginationEl.classList.add('hidden');
      return;
    }

    resultsGrid.classList.remove('hidden');
    if (emptyEl) emptyEl.classList.add('hidden');

    const totalPages = Math.ceil(this.filteredProfiles.length / this.pageSize);
    const pageItems = this.getCurrentPageItems();

    resultsGrid.innerHTML = pageItems.map((p) => {
      const profileName = p.name || 'N/A';
      const initials = UIUtils.getProfileInitials(profileName);
      const avatarColor = UIUtils.getAvatarColor(profileName);

      const rateLimit = p.rateLimit || p.rate_limit || 'N/A';
      const localAddress = p.localAddress || p.local_address || '-';
      const remoteAddress = p.remoteAddress || p.remote_address || '-';
      const parentQueue = p.parentQueue || p.parent_queue || '-';
      const dnsServer = p.dnsServer || p.dns_server || '';

      return `
        <div class="bg-white rounded-xl shadow-xs hover:shadow-md transition-all duration-200 border border-slate-200/80 p-4" data-name="${profileName}">
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
                  ${profileName}
                </h4>
                <span class="inline-flex items-center justify-center px-2.5 py-0.5 rounded-full text-xs font-semibold border bg-emerald-50 text-emerald-700 border-emerald-200">
                  <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="mr-1 text-emerald-600 inline-block"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/></svg>
                  <span class="hidden sm:inline">Aktif</span>
                  <span class="sm:hidden">On</span>
                </span>
              </div>
              
              <!-- Info Details -->
              <div class="flex flex-col gap-1.5 text-xs text-slate-500">
                <p class="text-sm text-slate-700 flex items-center font-medium truncate">
                  <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="mr-1.5 text-slate-400 shrink-0"><path d="m12 14 4-4"/><path d="M3.34 19a10 10 0 1 1 17.32 0"/></svg>
                  <span class="font-mono text-blue-700 font-semibold">${rateLimit}</span>
                </p>
                
                <div class="grid grid-cols-1 sm:grid-cols-2 gap-x-4 gap-y-1 text-xs text-slate-500">
                  <p class="truncate"><span class="text-slate-400">Local:</span> <span class="font-mono text-slate-700">${localAddress}</span></p>
                  <p class="truncate"><span class="text-slate-400">Remote:</span> <span class="font-mono text-slate-700">${remoteAddress}</span></p>
                  <p class="truncate"><span class="text-slate-400">Queue:</span> <span class="text-slate-700">${parentQueue}</span></p>
                  ${dnsServer ? `<p class="truncate"><span class="text-slate-400">DNS:</span> <span class="font-mono text-slate-700">${dnsServer}</span></p>` : ''}
                </div>
              </div>
            </div>

            <!-- Desktop Actions -->
            <div class="hidden sm:flex flex-col gap-2 items-end justify-center pl-4 border-l border-slate-100">
              <button onclick="PPPProfiles.editProfile('${profileName}')" class="p-2 bg-white border border-blue-200 text-blue-600 hover:bg-blue-50 rounded-lg transition-all shadow-xs cursor-pointer" title="Edit Profile">
                <svg xmlns="http://www.w3.org/2000/svg" width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M17 3a2.85 2.83 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5Z"/><path d="m15 5 4 4"/></svg>
              </button>
            </div>
            
            <!-- Mobile: Three-dot dropdown menu -->
            <div class="sm:hidden relative">
              <button class="action-menu-btn p-2 text-slate-500 hover:bg-slate-100 rounded-full transition-colors cursor-pointer" type="button" aria-label="Menu Opsi">
                <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="1"/><circle cx="12" cy="5" r="1"/><circle cx="12" cy="19" r="1"/></svg>
              </button>
              <div class="action-menu hidden absolute right-0 mt-1 w-44 bg-white rounded-xl shadow-xl border border-slate-100 overflow-hidden z-50">
                <button onclick="PPPProfiles.editProfile('${profileName}')" class="w-full text-left px-4 py-3 text-sm text-blue-600 hover:bg-blue-50 flex items-center gap-3 cursor-pointer">
                  <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M17 3a2.85 2.83 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5Z"/><path d="m15 5 4 4"/></svg>
                  <span>Edit Profile</span>
                </button>
              </div>
            </div>

          </div>
        </div>
      `;
    }).join('');

    this.renderPagination(totalPages);
  },

  editProfile(profileName) {
    showToast(`Edit profil ${profileName} dapat dilakukan via terminal MikroTik atau menu konfigurasi`, 'info');
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
    const total = this.filteredProfiles.length;
    const startIdx = (this.currentPage - 1) * this.pageSize + 1;
    const endIdx = Math.min(this.currentPage * this.pageSize, total);

    el.innerHTML = `
      <div class="flex flex-col sm:flex-row justify-between items-center gap-4 text-xs text-slate-500 border-t border-slate-200/80 pt-4">
        <p>Menampilkan <span class="font-semibold text-slate-700">${startIdx}</span> - <span class="font-semibold text-slate-700">${endIdx}</span> dari <span class="font-semibold text-slate-700">${total}</span> profil</p>
        <div class="flex items-center gap-1.5">
          <button ${this.currentPage <= 1 ? 'disabled' : ''} onclick="PPPProfiles.goToPage(${this.currentPage - 1})" class="px-3 py-1.5 rounded-lg border border-slate-300 bg-white hover:bg-slate-50 text-slate-700 disabled:opacity-40 disabled:cursor-not-allowed shadow-xs cursor-pointer">Sebelumnya</button>
          <span class="px-2 text-slate-700 font-semibold">${this.currentPage} / ${totalPages}</span>
          <button ${this.currentPage >= totalPages ? 'disabled' : ''} onclick="PPPProfiles.goToPage(${this.currentPage + 1})" class="px-3 py-1.5 rounded-lg border border-slate-300 bg-white hover:bg-slate-50 text-slate-700 disabled:opacity-40 disabled:cursor-not-allowed shadow-xs cursor-pointer">Berikutnya</button>
        </div>
      </div>
    `;
  },

  goToPage(page) {
    this.currentPage = page;
    this.render();
  }
};

window.PPPProfiles = PPPProfiles;
