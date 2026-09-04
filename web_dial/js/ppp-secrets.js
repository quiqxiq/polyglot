/**
 * Polyglot Web Dial PPP Secrets Controller
 */

const PPPSecrets = {
  currentDeviceId: null,
  allSecrets: [],
  filteredSecrets: [],
  currentPage: 1,
  pageSize: 25,

  init() {
    RouterSelector.onDeviceChange((deviceId) => {
      if (!deviceId) return;
      this.currentDeviceId = deviceId;
      this.loadSecrets();
    });

    this.bindControls();
  },

  async loadSecrets() {
    const loadingEl = document.getElementById('loading-state');
    const resultsGrid = document.getElementById('results-grid');
    const emptyEl = document.getElementById('empty-state');
    const paginationEl = document.getElementById('pagination-container');

    if (loadingEl) loadingEl.classList.remove('hidden');
    if (resultsGrid) resultsGrid.classList.add('hidden');
    if (emptyEl) emptyEl.classList.add('hidden');
    if (paginationEl) paginationEl.classList.add('hidden');

    try {
      const res = await API.call('/polyglot.v1.PPPService/ListSecrets', {
        device_id: this.currentDeviceId
      });

      this.allSecrets = res.secrets || [];
      this.populateProfileFilter();
      this.applyFilter();
    } catch (err) {
      console.error('Failed to load secrets:', err);
      showToast('Gagal memuat data PPP Secrets: ' + err.message, 'error');
    } finally {
      if (loadingEl) loadingEl.classList.add('hidden');
    }
  },

  populateProfileFilter() {
    const select = document.getElementById('profile-filter');
    if (!select) return;

    const currentVal = select.value;
    const profiles = Array.from(new Set(this.allSecrets.map((s) => s.profile).filter(Boolean))).sort();

    select.innerHTML = '<option value="">Semua Profile</option>';
    profiles.forEach((p) => {
      const opt = document.createElement('option');
      opt.value = p;
      opt.textContent = p;
      if (p === currentVal) opt.selected = true;
      select.appendChild(opt);
    });
  },

  bindControls() {
    const searchInput = document.getElementById('search-input');
    const profileFilter = document.getElementById('profile-filter');
    const pageSizeSelect = document.getElementById('per-page-select');
    const refreshBtn = document.getElementById('refresh-btn');

    if (searchInput) {
      searchInput.addEventListener('input', () => {
        this.currentPage = 1;
        this.applyFilter();
      });
    }

    if (profileFilter) {
      profileFilter.addEventListener('change', () => {
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
      refreshBtn.onclick = () => this.loadSecrets();
    }
  },

  applyFilter() {
    const searchVal = (document.getElementById('search-input')?.value || '').toLowerCase().trim();
    const profileVal = document.getElementById('profile-filter')?.value || '';

    this.filteredSecrets = this.allSecrets.filter((s) => {
      const name = (s.name || '').toLowerCase();
      const matchSearch = !searchVal || name.includes(searchVal);
      const matchProfile = !profileVal || s.profile === profileVal;

      return matchSearch && matchProfile;
    });

    const countEl = document.getElementById('results-count');
    if (countEl) countEl.textContent = `(${this.filteredSecrets.length} akun)`;

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
      const profile = s.profile || 'none';
      const disabled = s.disabled === true || s.disabled === 'true';
      const rosId = s.id || '';
      const service = s.service || 'any';

      // Status styling matching gnet_dial
      const statusClass = disabled ? 'bg-rose-50 text-rose-700 border-rose-200' : 'bg-emerald-50 text-emerald-700 border-emerald-200';
      const statusIcon = disabled
        ? '<svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="mr-1 text-rose-600 inline-block"><circle cx="12" cy="12" r="10"/><path d="m15 9-6 6"/><path d="m9 9 6 6"/></svg>'
        : '<svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="mr-1 text-emerald-600 inline-block"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/></svg>';
      const statusTextFull = disabled ? 'Disabled' : 'Enabled';
      const statusTextShort = disabled ? 'Off' : 'On';

      const toggleTitle = disabled ? 'Aktifkan Akun (Enable)' : 'Nonaktifkan Akun (Disable)';
      const toggleBorderClass = disabled
        ? 'border-emerald-200 text-emerald-600 hover:bg-emerald-50'
        : 'border-amber-200 text-amber-600 hover:bg-amber-50';

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
                <span class="inline-flex items-center justify-center px-2.5 py-0.5 rounded-full text-xs font-semibold border ${statusClass}">
                  ${statusIcon}
                  <span class="hidden sm:inline">${statusTextFull}</span>
                  <span class="sm:hidden">${statusTextShort}</span>
                </span>
              </div>

              <!-- Info Details: Only Profile and Service -->
              <div class="flex flex-wrap items-center gap-x-4 gap-y-1.5 text-xs text-slate-500">
                <p class="flex items-center truncate">
                  <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-shield mr-1.5 text-slate-400 shrink-0"><path d="M20 13c0 5-3.5 7.5-7.66 8.95a1 1 0 0 1-.68-.05A9.5 9.5 0 0 1 4 13c0-5 3.5-7.5 7.66-8.95a1 1 0 0 1 .68.05A9.5 9.5 0 0 1 20 13Z"/></svg>
                  <span class="font-medium text-slate-700">${profile}</span>
                </p>
                <span class="text-slate-300 hidden sm:inline">&bull;</span>
                <span class="text-slate-500 capitalize">Service: <span class="font-mono text-slate-700">${service}</span></span>
              </div>
            </div>

            <!-- Desktop Actions -->
            <div class="hidden sm:flex flex-col gap-2 items-end justify-center pl-4 border-l border-slate-100">
              <div class="flex gap-2">
                <button onclick="PPPSecrets.toggleSecret('${rosId}', ${!disabled})" class="p-2 bg-white border ${toggleBorderClass} rounded-lg transition-all shadow-xs cursor-pointer" title="${toggleTitle}">
                  ${disabled 
                    ? '<svg xmlns="http://www.w3.org/2000/svg" width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 2v10"/><path d="M18.4 6.6a9 9 0 1 1-12.77.04"/></svg>'
                    : '<svg xmlns="http://www.w3.org/2000/svg" width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="4.93" y1="4.93" x2="19.07" y2="19.07"/></svg>'
                  }
                </button>
                <a href="ppp-active.html?search=${encodeURIComponent(userName)}" class="p-2 bg-white border border-blue-200 text-blue-600 hover:bg-blue-50 rounded-lg transition-all shadow-xs cursor-pointer inline-flex items-center justify-center" title="Cek Sesi Aktif">
                  <svg xmlns="http://www.w3.org/2000/svg" width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M5 12h14"/><path d="m12 5 7 7-7 7"/></svg>
                </a>
              </div>
            </div>

            <!-- Mobile: Three-dot dropdown menu -->
            <div class="sm:hidden relative">
              <button class="action-menu-btn p-2 text-slate-500 hover:bg-slate-100 rounded-full transition-colors cursor-pointer" type="button" aria-label="Menu Opsi">
                <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="1"/><circle cx="12" cy="5" r="1"/><circle cx="12" cy="19" r="1"/></svg>
              </button>
              <div class="action-menu hidden absolute right-0 mt-1 w-48 bg-white rounded-xl shadow-xl border border-slate-100 overflow-hidden z-50">
                <button onclick="PPPSecrets.toggleSecret('${rosId}', ${!disabled})" class="w-full text-left px-4 py-3 text-sm ${disabled ? 'text-emerald-600 hover:bg-emerald-50' : 'text-amber-600 hover:bg-amber-50'} flex items-center gap-3 border-b border-slate-50 cursor-pointer">
                  ${disabled
                    ? '<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 2v10"/><path d="M18.4 6.6a9 9 0 1 1-12.77.04"/></svg><span>Aktifkan Akun</span>'
                    : '<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="4.93" y1="4.93" x2="19.07" y2="19.07"/></svg><span>Nonaktifkan Akun</span>'
                  }
                </button>
                <a href="ppp-active.html?search=${encodeURIComponent(userName)}" class="w-full text-left px-4 py-3 text-sm text-blue-600 hover:bg-blue-50 flex items-center gap-3 cursor-pointer">
                  <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M5 12h14"/><path d="m12 5 7 7-7 7"/></svg>
                  <span>Cek Sesi Aktif</span>
                </a>
              </div>
            </div>

          </div>
        </div>
      `;
    }).join('');

    this.renderPagination(totalPages);
  },

  async toggleSecret(rosId, newDisabled) {
    try {
      await API.call('/polyglot.v1.PPPService/SetSecretDisabled', {
        device_id: this.currentDeviceId,
        ros_id: rosId,
        disabled: newDisabled
      });
      showToast(`Status akun secret berhasil diubah`, 'success');
      this.loadSecrets();
    } catch (err) {
      showToast('Gagal mengubah status secret: ' + err.message, 'error');
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
        <p>Menampilkan <span class="font-semibold text-slate-700">${startIdx}</span> - <span class="font-semibold text-slate-700">${endIdx}</span> dari <span class="font-semibold text-slate-700">${total}</span> akun secret</p>
        <div class="flex items-center gap-1.5">
          <button ${this.currentPage <= 1 ? 'disabled' : ''} onclick="PPPSecrets.goToPage(${this.currentPage - 1})" class="px-3 py-1.5 rounded-lg border border-slate-300 bg-white hover:bg-slate-50 text-slate-700 disabled:opacity-40 disabled:cursor-not-allowed shadow-xs cursor-pointer">Sebelumnya</button>
          <span class="px-2 text-slate-700 font-semibold">${this.currentPage} / ${totalPages}</span>
          <button ${this.currentPage >= totalPages ? 'disabled' : ''} onclick="PPPSecrets.goToPage(${this.currentPage + 1})" class="px-3 py-1.5 rounded-lg border border-slate-300 bg-white hover:bg-slate-50 text-slate-700 disabled:opacity-40 disabled:cursor-not-allowed shadow-xs cursor-pointer">Berikutnya</button>
        </div>
      </div>
    `;
  },

  goToPage(page) {
    this.currentPage = page;
    this.render();
  }
};

window.PPPSecrets = PPPSecrets;
