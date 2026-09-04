/**
 * Polyglot Web Dial PPP Profiles Controller
 */

const PPPProfiles = {
  currentDeviceId: null,
  allProfiles: [],
  filteredProfiles: [],

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
    const tableEl = document.getElementById('profiles-table');
    const emptyEl = document.getElementById('empty-state');

    if (loadingEl) loadingEl.classList.remove('hidden');
    if (tableEl) tableEl.classList.add('hidden');
    if (emptyEl) emptyEl.classList.add('hidden');

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
    const refreshBtn = document.getElementById('refresh-btn');

    if (searchInput) {
      searchInput.addEventListener('input', () => this.applyFilter());
    }
    if (refreshBtn) {
      refreshBtn.onclick = () => this.loadProfiles();
    }
  },

  applyFilter() {
    const query = (document.getElementById('search-input')?.value || '').toLowerCase().trim();

    this.filteredProfiles = this.allProfiles.filter(p => {
      if (!query) return true;
      return (p.name && p.name.toLowerCase().includes(query)) ||
             (p.rate_limit && p.rate_limit.toLowerCase().includes(query)) ||
             (p.local_address && p.local_address.toLowerCase().includes(query)) ||
             (p.remote_address && p.remote_address.toLowerCase().includes(query));
    });

    const countEl = document.getElementById('results-count');
    if (countEl) countEl.textContent = `(${this.filteredProfiles.length} profil)`;

    this.renderTable();
  },

  renderTable() {
    const tbody = document.getElementById('profiles-tbody');
    const tableEl = document.getElementById('profiles-table');
    const emptyEl = document.getElementById('empty-state');
    if (!tbody) return;

    if (this.filteredProfiles.length === 0) {
      if (tableEl) tableEl.classList.add('hidden');
      if (emptyEl) emptyEl.classList.remove('hidden');
      return;
    }

    if (tableEl) tableEl.classList.remove('hidden');
    if (emptyEl) emptyEl.classList.add('hidden');

    tbody.innerHTML = this.filteredProfiles.map((p, idx) => {
      return `
        <tr class="hover:bg-slate-50 transition-colors border-b border-slate-100 text-sm">
          <td class="py-3 px-4 text-slate-400 font-mono text-xs">${idx + 1}</td>
          <td class="py-3 px-4 font-bold text-slate-900">${p.name}</td>
          <td class="py-3 px-4 font-mono text-xs text-blue-700 font-semibold">${p.rate_limit || 'Unlimited'}</td>
          <td class="py-3 px-4 text-xs font-mono text-slate-600">${p.local_address || '-'}</td>
          <td class="py-3 px-4 text-xs font-mono text-slate-600">${p.remote_address || '-'}</td>
          <td class="py-3 px-4 text-xs text-slate-500">${p.dns_server || '-'}</td>
          <td class="py-3 px-4 text-xs text-slate-500">${p.parent_queue || '-'}</td>
          <td class="py-3 px-4 text-right">
            <span class="px-2.5 py-0.5 rounded-full text-xs font-semibold ${p.only_one === 'yes' ? 'bg-amber-50 text-amber-700' : 'bg-slate-100 text-slate-600'}">
              ${p.only_one || 'default'}
            </span>
          </td>
        </tr>
      `;
    }).join('');
  }
};

window.PPPProfiles = PPPProfiles;
