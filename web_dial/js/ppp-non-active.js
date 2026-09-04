/**
 * Polyglot Web Dial PPP Non-Active (Offline) Controller
 */

const PPPNonActive = {
  currentDeviceId: null,
  allSecrets: [],
  filteredSecrets: [],
  currentPage: 1,
  pageSize: 25,

  init() {
    RouterSelector.onDeviceChange((deviceId) => {
      if (!deviceId) return;
      this.currentDeviceId = deviceId;
      this.loadInactive();
    });

    this.bindControls();
  },

  async loadInactive() {
    const loadingEl = document.getElementById('loading-state');
    const tableEl = document.getElementById('inactive-table');
    const emptyEl = document.getElementById('empty-state');

    if (loadingEl) loadingEl.classList.remove('hidden');
    if (tableEl) tableEl.classList.add('hidden');
    if (emptyEl) emptyEl.classList.add('hidden');

    try {
      const res = await API.call('/polyglot.v1.PPPService/ListInactiveSecrets', {
        device_id: this.currentDeviceId
      });

      this.allSecrets = res.secrets || [];
      this.applyFilter();
    } catch (err) {
      console.error('Failed to load inactive secrets:', err);
      showToast('Gagal memuat pelanggan offline: ' + err.message, 'error');
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
        this.renderTable();
      });
    }

    if (refreshBtn) {
      refreshBtn.onclick = () => this.loadInactive();
    }
  },

  applyFilter() {
    const query = (document.getElementById('search-input')?.value || '').toLowerCase().trim();

    this.filteredSecrets = this.allSecrets.filter(s => {
      if (!query) return true;
      return (s.name && s.name.toLowerCase().includes(query)) ||
             (s.comment && s.comment.toLowerCase().includes(query)) ||
             (s.remote_address && s.remote_address.toLowerCase().includes(query)) ||
             (s.profile && s.profile.toLowerCase().includes(query));
    });

    const countEl = document.getElementById('results-count');
    if (countEl) countEl.textContent = `(${this.filteredSecrets.length} offline)`;

    this.renderTable();
  },

  renderTable() {
    const tbody = document.getElementById('inactive-tbody');
    const tableEl = document.getElementById('inactive-table');
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
      return `
        <tr class="hover:bg-slate-50 transition-colors border-b border-slate-100 text-sm">
          <td class="py-3 px-4 text-slate-400 font-mono text-xs">${startIdx + idx + 1}</td>
          <td class="py-3 px-4">
            <span class="font-bold text-slate-900">${s.name}</span>
            <span class="block text-[11px] text-slate-400 font-mono">${s.caller_id || 'no-mac'}</span>
          </td>
          <td class="py-3 px-4">
            <span class="px-2.5 py-0.5 rounded-full text-xs font-semibold bg-slate-100 text-slate-700">
              ${s.profile || 'default'}
            </span>
          </td>
          <td class="py-3 px-4 text-slate-600 text-xs font-mono">${s.remote_address || '-'}</td>
          <td class="py-3 px-4 text-slate-500 text-xs">${s.last_logged_out || '-'}</td>
          <td class="py-3 px-4 text-slate-500 text-xs max-w-xs truncate">${s.comment || '-'}</td>
          <td class="py-3 px-4 text-right">
            <span class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-semibold bg-slate-100 text-slate-600">
              Offline
            </span>
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
          <button ${this.currentPage <= 1 ? 'disabled' : ''} onclick="PPPNonActive.goToPage(${this.currentPage - 1})" class="px-3 py-1.5 rounded-lg border border-slate-200 hover:bg-slate-100 disabled:opacity-40">Sebelumnya</button>
          <button ${this.currentPage >= totalPages ? 'disabled' : ''} onclick="PPPNonActive.goToPage(${this.currentPage + 1})" class="px-3 py-1.5 rounded-lg border border-slate-200 hover:bg-slate-100 disabled:opacity-40">Berikutnya</button>
        </div>
      </div>
    `;
  },

  goToPage(page) {
    this.currentPage = page;
    this.renderTable();
  }
};

window.PPPNonActive = PPPNonActive;
