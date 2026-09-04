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
    const tableEl = document.getElementById('secrets-table');
    const emptyEl = document.getElementById('empty-state');

    if (loadingEl) loadingEl.classList.remove('hidden');
    if (tableEl) tableEl.classList.add('hidden');
    if (emptyEl) emptyEl.classList.add('hidden');

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
    const profiles = Array.from(new Set(this.allSecrets.map(s => s.profile).filter(Boolean))).sort();

    select.innerHTML = '<option value="">Semua Profile</option>';
    profiles.forEach(p => {
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
        this.renderTable();
      });
    }

    if (refreshBtn) {
      refreshBtn.onclick = () => this.loadSecrets();
    }
  },

  applyFilter() {
    const searchVal = (document.getElementById('search-input')?.value || '').toLowerCase().trim();
    const profileVal = document.getElementById('profile-filter')?.value || '';

    this.filteredSecrets = this.allSecrets.filter(s => {
      const matchSearch = !searchVal ||
        (s.name && s.name.toLowerCase().includes(searchVal)) ||
        (s.comment && s.comment.toLowerCase().includes(searchVal)) ||
        (s.remote_address && s.remote_address.toLowerCase().includes(searchVal));

      const matchProfile = !profileVal || s.profile === profileVal;

      return matchSearch && matchProfile;
    });

    const countEl = document.getElementById('results-count');
    if (countEl) countEl.textContent = `(${this.filteredSecrets.length} akun)`;

    this.renderTable();
  },

  renderTable() {
    const tbody = document.getElementById('secrets-tbody');
    const tableEl = document.getElementById('secrets-table');
    const emptyEl = document.getElementById('empty-state');
    const paginationEl = document.getElementById('pagination-container');
    if (!tbody) return;

    if (this.filteredSecrets.length === 0) {
      if (tableEl) tableEl.classList.add('hidden');
      if (emptyEl) emptyEl.classList.remove('hidden');
      if (paginationEl) paginationEl.innerHTML = '';
      return;
    }

    if (tableEl) tableEl.classList.remove('hidden');
    if (emptyEl) emptyEl.classList.add('hidden');

    const totalPages = Math.ceil(this.filteredSecrets.length / this.pageSize);
    const startIdx = (this.currentPage - 1) * this.pageSize;
    const pageItems = this.filteredSecrets.slice(startIdx, startIdx + this.pageSize);

    tbody.innerHTML = pageItems.map((s, idx) => {
      const isDisabled = s.disabled;
      return `
        <tr class="hover:bg-slate-50 transition-colors border-b border-slate-100 text-sm">
          <td class="py-3 px-4 text-slate-400 font-mono text-xs">${startIdx + idx + 1}</td>
          <td class="py-3 px-4 font-semibold text-slate-900">
            <div class="flex items-center gap-2">
              <span>${s.name}</span>
              ${isDisabled ? '<span class="text-[10px] bg-rose-100 text-rose-700 px-1.5 py-0.2 rounded font-bold">DISABLED</span>' : ''}
            </div>
          </td>
          <td class="py-3 px-4">
            <span class="px-2.5 py-0.5 rounded-full text-xs font-semibold bg-blue-50 text-blue-700 border border-blue-200">
              ${s.profile || 'default'}
            </span>
          </td>
          <td class="py-3 px-4 text-slate-600 text-xs">${s.service || 'any'}</td>
          <td class="py-3 px-4 text-slate-600 text-xs font-mono">${s.remote_address || '-'}</td>
          <td class="py-3 px-4 text-slate-500 text-xs max-w-xs truncate">${s.comment || '-'}</td>
          <td class="py-3 px-4 text-right">
            <a href="ppp-active.html?search=${encodeURIComponent(s.name)}" class="text-xs text-blue-600 hover:underline font-medium">
              Cek Status &rarr;
            </a>
          </td>
        </tr>
      `;
    }).join('');

    this.renderPagination(totalPages);
  },

  renderPagination(totalPages) {
    const el = document.getElementById('pagination-container');
    if (!el || totalPages <= 1) {
      if (el) el.innerHTML = '';
      return;
    }

    el.innerHTML = `
      <div class="flex items-center justify-between py-3 border-t border-slate-200 text-xs text-slate-500">
        <span>Halaman ${this.currentPage} dari ${totalPages}</span>
        <div class="flex gap-1.5">
          <button ${this.currentPage <= 1 ? 'disabled' : ''} onclick="PPPSecrets.goToPage(${this.currentPage - 1})" class="px-3 py-1.5 rounded-lg border border-slate-200 hover:bg-slate-100 disabled:opacity-40">Sebelumnya</button>
          <button ${this.currentPage >= totalPages ? 'disabled' : ''} onclick="PPPSecrets.goToPage(${this.currentPage + 1})" class="px-3 py-1.5 rounded-lg border border-slate-200 hover:bg-slate-100 disabled:opacity-40">Berikutnya</button>
        </div>
      </div>
    `;
  },

  goToPage(page) {
    this.currentPage = page;
    this.renderTable();
  }
};

window.PPPSecrets = PPPSecrets;
