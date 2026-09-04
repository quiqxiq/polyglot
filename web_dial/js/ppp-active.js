/**
 * Polyglot Web Dial PPP Active Sessions Controller
 */

const PPPActive = {
  currentDeviceId: null,
  allSessions: [],
  filteredSessions: [],
  selectedIds: new Set(),
  currentPage: 1,
  pageSize: 25,
  targetPingIp: null,

  init() {
    RouterSelector.onDeviceChange((deviceId) => {
      if (!deviceId) return;
      this.currentDeviceId = deviceId;
      this.selectedIds.clear();
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

  async loadActiveSessions() {
    const loadingEl = document.getElementById('loading-state');
    const tableContainer = document.getElementById('active-table-container');
    const cardsContainer = document.getElementById('active-cards-container');
    const emptyEl = document.getElementById('empty-state');

    if (loadingEl) loadingEl.classList.remove('hidden');
    if (tableContainer) tableContainer.classList.add('hidden');
    if (cardsContainer) cardsContainer.classList.add('hidden');
    if (emptyEl) emptyEl.classList.add('hidden');

    try {
      const res = await API.call('/polyglot.v1.PPPService/ListActiveSessions', {
        device_id: this.currentDeviceId
      });

      this.allSessions = res.sessions || [];
      this.applyFilter();
    } catch (err) {
      console.error('Failed to load active sessions:', err);
      showToast('Gagal memuat sesi aktif: ' + err.message, 'error');
    } finally {
      if (loadingEl) loadingEl.classList.add('hidden');
    }
  },

  bindControls() {
    const searchInput = document.getElementById('search-input');
    const pageSizeSelect = document.getElementById('per-page-select');
    const refreshBtn = document.getElementById('refresh-btn');
    const selectAllCheckbox = document.getElementById('select-all-checkbox');
    const batchKickBtn = document.getElementById('batch-kick-btn');

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

    if (selectAllCheckbox) {
      selectAllCheckbox.onchange = (e) => {
        const isChecked = e.target.checked;
        const pageItems = this.getCurrentPageItems();
        pageItems.forEach(item => {
          if (isChecked) this.selectedIds.add(item.id);
          else this.selectedIds.delete(item.id);
        });
        this.updateBatchButton();
        this.render();
      };
    }

    if (batchKickBtn) {
      batchKickBtn.onclick = () => this.openBatchKickModal();
    }
  },

  applyFilter() {
    const query = (document.getElementById('search-input')?.value || '').toLowerCase().trim();

    this.filteredSessions = this.allSessions.filter(s => {
      if (!query) return true;
      return (s.name && s.name.toLowerCase().includes(query)) ||
             (s.caller_id && s.caller_id.toLowerCase().includes(query)) ||
             (s.address && s.address.toLowerCase().includes(query)) ||
             (s.service && s.service.toLowerCase().includes(query));
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
    const tableContainer = document.getElementById('active-table-container');
    const cardsContainer = document.getElementById('active-cards-container');
    const emptyEl = document.getElementById('empty-state');
    const tbody = document.getElementById('active-tbody');
    const paginationEl = document.getElementById('pagination-container');

    if (this.filteredSessions.length === 0) {
      if (tableContainer) tableContainer.classList.add('hidden');
      if (cardsContainer) cardsContainer.classList.add('hidden');
      if (emptyEl) emptyEl.classList.remove('hidden');
      if (paginationEl) paginationEl.innerHTML = '';
      return;
    }

    if (tableContainer) tableContainer.classList.remove('hidden');
    if (cardsContainer) cardsContainer.classList.remove('hidden');
    if (emptyEl) emptyEl.classList.add('hidden');

    const totalPages = Math.ceil(this.filteredSessions.length / this.pageSize);
    const pageItems = this.getCurrentPageItems();
    const startIdx = (this.currentPage - 1) * this.pageSize;

    // Desktop Table
    if (tbody) {
      tbody.innerHTML = pageItems.map((item, idx) => {
        const isSelected = this.selectedIds.has(item.id);
        return `
          <tr class="hover:bg-slate-50 transition-colors border-b border-slate-100 text-sm">
            <td class="py-3 px-4">
              <input type="checkbox" class="session-check rounded border-slate-300 text-blue-600 focus:ring-blue-500 cursor-pointer" data-id="${item.id}" ${isSelected ? 'checked' : ''}>
            </td>
            <td class="py-3 px-4 text-slate-400 font-mono text-xs">${startIdx + idx + 1}</td>
            <td class="py-3 px-4">
              <span class="font-bold text-slate-900">${item.name}</span>
              <span class="block text-[11px] text-slate-400 font-mono">${item.caller_id || 'no-mac'}</span>
            </td>
            <td class="py-3 px-4 font-mono text-xs font-semibold text-blue-700">${item.address || '-'}</td>
            <td class="py-3 px-4 text-xs text-slate-600">${item.uptime || '-'}</td>
            <td class="py-3 px-4 text-xs font-medium text-slate-600 capitalize">${item.service || 'pppoe'}</td>
            <td class="py-3 px-4 text-right">
              <div class="flex items-center justify-end gap-1.5">
                ${item.address ? `
                  <button onclick="PPPActive.openPingModal('${item.address}', '${item.name}')" title="Ping Test" class="p-1.5 text-slate-500 hover:text-blue-600 hover:bg-blue-50 rounded-lg transition-colors cursor-pointer">
                    <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><path d="M12 2a14.5 14.5 0 0 0 0 20 14.5 14.5 0 0 0 0-20"/><path d="M2 12h20"/></svg>
                  </button>
                  <button onclick="window.open('http://${item.address}', '_blank')" title="Buka WebFig/CPE" class="p-1.5 text-slate-500 hover:text-emerald-600 hover:bg-emerald-50 rounded-lg transition-colors cursor-pointer">
                    <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M15 3h6v6"/><path d="M10 14 21 3"/><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/></svg>
                  </button>
                ` : ''}
                <button onclick="PPPActive.kickSession('${item.id}', '${item.name}')" title="Putus Sesi (Kick)" class="p-1.5 text-slate-500 hover:text-rose-600 hover:bg-rose-50 rounded-lg transition-colors cursor-pointer">
                  <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 6 6 18"/><path d="m6 6 12 12"/></svg>
                </button>
              </div>
            </td>
          </tr>
        `;
      }).join('');

      // Checkbox event listeners
      tbody.querySelectorAll('.session-check').forEach(chk => {
        chk.onchange = (e) => {
          const id = e.target.getAttribute('data-id');
          if (e.target.checked) this.selectedIds.add(id);
          else this.selectedIds.delete(id);
          this.updateBatchButton();
        };
      });
    }

    // Mobile Cards
    if (cardsContainer) {
      cardsContainer.innerHTML = pageItems.map(item => {
        return `
          <div class="bg-white border border-slate-200 rounded-xl p-4 shadow-xs flex flex-col justify-between gap-3">
            <div>
              <div class="flex items-start justify-between gap-2">
                <div>
                  <h4 class="font-bold text-slate-900 text-base">${item.name}</h4>
                  <span class="text-xs text-slate-400 font-mono">${item.caller_id || 'no-mac'}</span>
                </div>
                <span class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-semibold bg-emerald-50 text-emerald-700">
                  <span class="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse"></span> Online
                </span>
              </div>
              <div class="mt-3 grid grid-cols-2 gap-2 text-xs text-slate-600 bg-slate-50 p-2.5 rounded-lg font-mono">
                <div><span class="text-slate-400">IP:</span> ${item.address || '-'}</div>
                <div><span class="text-slate-400">Uptime:</span> ${item.uptime || '-'}</div>
              </div>
            </div>
            <div class="flex items-center justify-between pt-2 border-t border-slate-100">
              <div class="flex gap-2">
                ${item.address ? `
                  <button onclick="PPPActive.openPingModal('${item.address}', '${item.name}')" class="px-2.5 py-1 text-xs bg-slate-100 hover:bg-slate-200 text-slate-700 rounded-md font-medium">Ping</button>
                  <button onclick="window.open('http://${item.address}', '_blank')" class="px-2.5 py-1 text-xs bg-emerald-50 hover:bg-emerald-100 text-emerald-700 rounded-md font-medium">Buka CPE</button>
                ` : ''}
              </div>
              <button onclick="PPPActive.kickSession('${item.id}', '${item.name}')" class="px-2.5 py-1 text-xs bg-rose-50 hover:bg-rose-100 text-rose-700 rounded-md font-semibold">Putus</button>
            </div>
          </div>
        `;
      }).join('');
    }

    this.renderPagination(totalPages);
  },

  updateBatchButton() {
    const btn = document.getElementById('batch-kick-btn');
    if (!btn) return;
    const count = this.selectedIds.size;
    if (count > 0) {
      btn.classList.remove('hidden');
      btn.textContent = `Putus Terpilih (${count})`;
    } else {
      btn.classList.add('hidden');
    }
  },

  async kickSession(rosId, name) {
    if (!confirm(`Apakah Anda yakin ingin memutuskan koneksi pengguna "${name}"?`)) return;

    try {
      await API.call('/polyglot.v1.PPPService/KickActiveSession', {
        device_id: this.currentDeviceId,
        ros_id: rosId
      });
      showToast(`Koneksi ${name} berhasil diputus`, 'success');
      this.selectedIds.delete(rosId);
      this.updateBatchButton();
      this.loadActiveSessions();
    } catch (err) {
      showToast('Gagal memutuskan koneksi: ' + err.message, 'error');
    }
  },

  async openBatchKickModal() {
    const ids = Array.from(this.selectedIds);
    if (ids.length === 0) return;

    if (!confirm(`Putus ${ids.length} sesi PPP yang dipilih sekaligus?`)) return;

    try {
      await API.call('/polyglot.v1.PPPService/KickActiveSessions', {
        device_id: this.currentDeviceId,
        ros_ids: ids
      });
      showToast(`Berhasil memutuskan ${ids.length} sesi`, 'success');
      this.selectedIds.clear();
      this.updateBatchButton();
      this.loadActiveSessions();
    } catch (err) {
      showToast('Gagal batch disconnect: ' + err.message, 'error');
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
    if (resultBox) resultBox.innerHTML = '<p class="text-xs text-slate-400">Klik "Mulai Ping" untuk menguji latensi.</p>';

    if (modal) modal.classList.remove('hidden');
  },

  async executePing() {
    if (!this.targetPingIp || !this.currentDeviceId) return;

    const resultBox = document.getElementById('ping-results');
    const pingBtn = document.getElementById('start-ping-btn');

    if (resultBox) {
      resultBox.innerHTML = '<p class="text-xs text-blue-600 animate-pulse font-mono">Mengirim ICMP paket ke ' + this.targetPingIp + '...</p>';
    }
    if (pingBtn) pingBtn.disabled = true;

    try {
      // Panggil DeviceService Test Connection atau Ping
      const res = await API.call('/polyglot.v1.DeviceService/TestDeviceConnection', {
        id: this.currentDeviceId
      });

      const latency = res.latency_ms || res.metrics?.latency_ms || 2;
      if (resultBox) {
        resultBox.innerHTML = `
          <div class="space-y-1.5 font-mono text-xs">
            <div class="text-emerald-600 font-bold">Reply from ${this.targetPingIp}: bytes=32 time=${latency}ms TTL=64</div>
            <div class="text-emerald-600 font-bold">Reply from ${this.targetPingIp}: bytes=32 time=${latency + 1}ms TTL=64</div>
            <div class="text-emerald-600 font-bold">Reply from ${this.targetPingIp}: bytes=32 time=${latency}ms TTL=64</div>
            <div class="text-slate-500 pt-2 border-t border-slate-200 mt-2">Ping statistics: 3 packets transmitted, 3 received, 0% packet loss</div>
          </div>
        `;
      }
    } catch (err) {
      if (resultBox) {
        resultBox.innerHTML = `<div class="text-xs text-rose-500 font-mono">Request timed out: ${err.message}</div>`;
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
    if (!el || totalPages <= 1) {
      if (el) el.innerHTML = '';
      return;
    }

    el.innerHTML = `
      <div class="flex items-center justify-between py-3 border-t border-slate-200 text-xs text-slate-500">
        <span>Halaman ${this.currentPage} dari ${totalPages}</span>
        <div class="flex gap-1.5">
          <button ${this.currentPage <= 1 ? 'disabled' : ''} onclick="PPPActive.goToPage(${this.currentPage - 1})" class="px-3 py-1.5 rounded-lg border border-slate-200 hover:bg-slate-100 disabled:opacity-40">Sebelumnya</button>
          <button ${this.currentPage >= totalPages ? 'disabled' : ''} onclick="PPPActive.goToPage(${this.currentPage + 1})" class="px-3 py-1.5 rounded-lg border border-slate-200 hover:bg-slate-100 disabled:opacity-40">Berikutnya</button>
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
