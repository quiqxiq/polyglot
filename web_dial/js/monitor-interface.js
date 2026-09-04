/**
 * Polyglot Web Dial Interface & Simple Queue Monitor Controller
 * Arsitektur Hemat Sumber Daya:
 * - Menampilkan daftar semua interface dan simple queue (tanpa streaming background berlebih).
 * - Realtime RX/TX chart hanya streaming on-demand saat baris diklik dan modal terbuka.
 * - Mematikan stream segera saat modal ditutup.
 */

const MonitorInterface = {
  currentDeviceId: null,
  activeTab: 'interfaces', // 'interfaces' | 'queues'
  interfaces: [],
  queues: [],
  searchQuery: '',

  // State untuk Realtime Traffic Modal (On-Demand)
  trafficController: null,
  trafficHistory: [],
  historyLength: 60,
  peakRx: 0,
  peakTx: 0,

  init() {
    RouterSelector.onDeviceChange((deviceId) => {
      if (!deviceId) return;
      this.currentDeviceId = deviceId;
      this.loadData();
    });

    const refreshBtn = document.getElementById('refresh-btn');
    if (refreshBtn) refreshBtn.onclick = () => this.loadData();

    const searchInput = document.getElementById('search-input');
    const clearBtn = document.getElementById('clear-search-btn');
    if (searchInput) {
      searchInput.oninput = (e) => {
        this.searchQuery = (e.target.value || '').trim().toLowerCase();
        if (clearBtn) clearBtn.classList.toggle('hidden', !this.searchQuery);
        this.renderActiveTable();
      };
    }
    if (clearBtn && searchInput) {
      clearBtn.onclick = () => {
        searchInput.value = '';
        this.searchQuery = '';
        clearBtn.classList.add('hidden');
        this.renderActiveTable();
      };
    }

    const modal = document.getElementById('traffic-modal');
    if (modal) {
      modal.onclick = (e) => { if (e.target === modal) this.closeTrafficModal(); };
    }
    window.addEventListener('keydown', (e) => {
      if (e.key === 'Escape' && modal && !modal.classList.contains('hidden')) {
        this.closeTrafficModal();
      }
    });
  },

  switchTab(tab) {
    this.activeTab = tab;
    const ifaceBtn = document.getElementById('tab-interfaces-btn');
    const queueBtn = document.getElementById('tab-queues-btn');
    const activeCls = 'px-4 py-1.5 rounded-lg text-xs font-bold transition-all bg-white text-slate-900 shadow-xs cursor-pointer flex items-center gap-1.5';
    const inactiveCls = 'px-4 py-1.5 rounded-lg text-xs font-semibold text-slate-500 hover:text-slate-900 transition-all cursor-pointer flex items-center gap-1.5';

    if (ifaceBtn) ifaceBtn.className = tab === 'interfaces' ? activeCls : inactiveCls;
    if (queueBtn) queueBtn.className = tab === 'queues' ? activeCls : inactiveCls;
    this.renderActiveTable();
  },

  async loadData(silent = false) {
    if (!this.currentDeviceId) return;

    const loadingEl = document.getElementById('loading-state');
    const ifaceTable = document.getElementById('interface-table');
    const queueTable = document.getElementById('queue-table');
    const emptyEl = document.getElementById('empty-state');
    const refreshIcon = document.getElementById('refresh-icon');

    if (!silent) {
      if (loadingEl) loadingEl.classList.remove('hidden');
      if (ifaceTable) ifaceTable.classList.add('hidden');
      if (queueTable) queueTable.classList.add('hidden');
      if (emptyEl) emptyEl.classList.add('hidden');
    }
    if (refreshIcon) refreshIcon.classList.add('animate-spin');

    try {
      await Promise.allSettled([this.loadInterfaces(), this.loadQueues()]);
      this.renderActiveTable();
    } catch (err) {
      console.error('Failed to load router data:', err);
      if (!silent) showToast('Gagal memuat data router: ' + err.message, 'error');
    } finally {
      if (loadingEl) loadingEl.classList.add('hidden');
      if (refreshIcon) refreshIcon.classList.remove('animate-spin');
    }
  },

  async loadInterfaces() {
    try {
      const res = await API.call('/polyglot.v1.DeviceService/TestDeviceConnection', {
        id: this.currentDeviceId,
        interfaceTypeFilter: ''
      });
      this.interfaces = res.interfaceList || res.interface_list || [];
      const countEl = document.getElementById('count-interfaces');
      if (countEl) countEl.textContent = this.interfaces.length;
    } catch (err) {
      console.warn('Failed to load interfaces:', err);
      this.interfaces = [];
    }
  },

  escapeHtml(str) {
    if (str === null || str === undefined) return '';
    return String(str)
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#039;');
  },

  async loadQueues() {
    try {
      this.queues = [];
      const queueMap = new Map();

      await new Promise((resolve) => {
        const controller = new AbortController();
        let debounceTimer = null;
        const maxTimeout = setTimeout(() => {
          controller.abort();
          resolve();
        }, 3500);

        API.stream(
          '/polyglot.v1.NetworkMonitorService/StreamQueueStats',
          { deviceId: this.currentDeviceId, interval: '1s' },
          (frame) => {
            const list = frame.queues || [];
            list.forEach(q => {
              if (q && q.name) {
                queueMap.set(q.id || q.name, q);
              }
            });
            if (debounceTimer) clearTimeout(debounceTimer);
            // Begitu tidak ada frame baru selama 250ms, snapshot burst selesai
            debounceTimer = setTimeout(() => {
              clearTimeout(maxTimeout);
              controller.abort();
              resolve();
            }, 250);
          },
          () => {
            clearTimeout(maxTimeout);
            if (debounceTimer) clearTimeout(debounceTimer);
            resolve();
          },
          controller.signal
        );
      });

      this.queues = Array.from(queueMap.values());
      const countEl = document.getElementById('count-queues');
      if (countEl) countEl.textContent = this.queues.length;
    } catch (err) {
      console.warn('Failed to load queues snapshot:', err);
      this.queues = [];
    }
  },

  renderActiveTable() {
    const ifaceTable = document.getElementById('interface-table');
    const queueTable = document.getElementById('queue-table');
    if (this.activeTab === 'interfaces') {
      if (queueTable) queueTable.classList.add('hidden');
      this.renderInterfacesTable();
    } else {
      if (ifaceTable) ifaceTable.classList.add('hidden');
      this.renderQueuesTable();
    }
  },

  renderInterfacesTable() {
    const tbody = document.getElementById('interface-tbody');
    const ifaceTable = document.getElementById('interface-table');
    const emptyEl = document.getElementById('empty-state');
    if (!tbody) return;

    let items = this.interfaces;
    if (this.searchQuery) {
      items = items.filter(i => 
        (i.name && i.name.toLowerCase().includes(this.searchQuery)) ||
        (i.type && i.type.toLowerCase().includes(this.searchQuery)) ||
        (i.macAddress && i.macAddress.toLowerCase().includes(this.searchQuery))
      );
    }

    if (items.length === 0) {
      if (ifaceTable) ifaceTable.classList.add('hidden');
      if (emptyEl) {
        emptyEl.classList.remove('hidden');
        document.getElementById('empty-state-msg').textContent = this.searchQuery 
          ? 'Tidak ada interface yang cocok dengan kata kunci pencarian.'
          : 'Router ini belum memiliki interface yang terdeteksi.';
      }
      return;
    }

    if (ifaceTable) ifaceTable.classList.remove('hidden');
    if (emptyEl) emptyEl.classList.add('hidden');

    tbody.innerHTML = items.map((iface, idx) => {
      const isRunning = Boolean(iface.running);
      const isDisabled = Boolean(iface.disabled);

      let statusBadge = '<span class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-semibold bg-slate-100 text-slate-600">Link Down</span>';
      if (isDisabled) {
        statusBadge = '<span class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-semibold bg-rose-50 text-rose-700 border border-rose-200">Disabled</span>';
      } else if (isRunning) {
        statusBadge = '<span class="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-emerald-50 text-emerald-700 border border-emerald-200"><span class="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse"></span> Running</span>';
      }

      const encodedName = encodeURIComponent(iface.name || '');
      const encodedSubtitle = encodeURIComponent(`Type: ${iface.type || 'ether'} &bull; MAC: ${iface.macAddress || iface.mac_address || '-'}`);

      return `
        <tr class="hover:bg-slate-50 transition-colors border-b border-slate-100 text-sm">
          <td class="py-3 px-4 font-mono text-xs text-slate-400">${idx + 1}</td>
          <td class="py-3 px-4 font-bold text-slate-900 font-mono">${this.escapeHtml(iface.name)}</td>
          <td class="py-3 px-4"><span class="inline-block px-2 py-0.5 rounded text-[11px] font-semibold uppercase bg-slate-100 text-slate-700">${this.escapeHtml(iface.type || 'ether')}</span></td>
          <td class="py-3 px-4">${statusBadge}</td>
          <td class="py-3 px-4 font-mono text-xs text-slate-500">${this.escapeHtml(iface.macAddress || iface.mac_address || '-')}</td>
          <td class="py-3 px-4 text-right">
            <button onclick="MonitorInterface.openTrafficModal('interface', decodeURIComponent('${encodedName}'), decodeURIComponent('${encodedSubtitle}'), ${isRunning})" class="inline-flex items-center gap-1.5 px-3 py-1.5 bg-blue-50 hover:bg-blue-100 text-blue-700 rounded-xl text-xs font-bold transition-colors cursor-pointer">
              <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/></svg>
              Monitor Traffic
            </button>
          </td>
        </tr>
      `;
    }).join('');
  },

  renderQueuesTable() {
    const tbody = document.getElementById('queue-tbody');
    const queueTable = document.getElementById('queue-table');
    const emptyEl = document.getElementById('empty-state');
    if (!tbody) return;

    let items = this.queues;
    if (this.searchQuery) {
      items = items.filter(q => 
        (q.name && q.name.toLowerCase().includes(this.searchQuery)) ||
        (q.target && q.target.toLowerCase().includes(this.searchQuery)) ||
        (q.maxLimit && q.maxLimit.toLowerCase().includes(this.searchQuery))
      );
    }

    if (items.length === 0) {
      if (queueTable) queueTable.classList.add('hidden');
      if (emptyEl) {
        emptyEl.classList.remove('hidden');
        document.getElementById('empty-state-msg').textContent = this.searchQuery 
          ? 'Tidak ada antrean yang cocok dengan kata kunci pencarian.'
          : 'Belum ada Simple Queue yang terkonfigurasi pada router ini (RouterOS /queue/simple kosong).';
      }
      return;
    }

    if (queueTable) queueTable.classList.remove('hidden');
    if (emptyEl) emptyEl.classList.add('hidden');

    tbody.innerHTML = items.map((q, idx) => {
      const isDisabled = Boolean(q.disabled);
      const targetStr = q.target || '-';
      const maxLimitStr = q.maxLimit || q.max_limit || 'unlimited';
      const bytesStr = q.bytes && q.bytes.includes('/')
        ? `${this.formatBytes(q.bytes.split('/')[0])} / ${this.formatBytes(q.bytes.split('/')[1])}`
        : this.formatBytes(q.bytes);
      const statusBadge = isDisabled
        ? '<span class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-semibold bg-slate-100 text-slate-500">Disabled</span>'
        : '<span class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full text-xs font-semibold bg-emerald-50 text-emerald-700 border border-emerald-200"><span class="w-1.5 h-1.5 rounded-full bg-emerald-500"></span> Aktif</span>';

      const encodedName = encodeURIComponent(q.name || '');
      const encodedSubtitle = encodeURIComponent(`Target: ${targetStr} &bull; Max: ${maxLimitStr}`);

      return `
        <tr class="hover:bg-slate-50 transition-colors border-b border-slate-100 text-sm">
          <td class="py-3 px-4 font-mono text-xs text-slate-400">${idx + 1}</td>
          <td class="py-3 px-4 font-bold text-slate-900 font-mono">${this.escapeHtml(q.name)}</td>
          <td class="py-3 px-4 font-mono text-xs text-blue-600">${this.escapeHtml(targetStr)}</td>
          <td class="py-3 px-4 font-mono text-xs text-slate-700">${this.escapeHtml(maxLimitStr)}</td>
          <td class="py-3 px-4 font-mono text-xs text-slate-600" title="${q.bytes || ''}">${bytesStr}</td>
          <td class="py-3 px-4">${statusBadge}</td>
          <td class="py-3 px-4 text-right">
            <button onclick="MonitorInterface.openTrafficModal('queue', decodeURIComponent('${encodedName}'), decodeURIComponent('${encodedSubtitle}'), ${!isDisabled})" class="inline-flex items-center gap-1.5 px-3 py-1.5 bg-indigo-50 hover:bg-indigo-100 text-indigo-700 rounded-xl text-xs font-bold transition-colors cursor-pointer">
              <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/></svg>
              Monitor Traffic
            </button>
          </td>
        </tr>
      `;
    }).join('');
  },

  // ── ON-DEMAND LIVE TRAFFIC MODAL SYSTEM ────────────────────────────────────

  openTrafficModal(type, targetName, subtitle, isRunning = true) {
    if (!this.currentDeviceId) return;
    this.closeTrafficModal();

    const modal = document.getElementById('traffic-modal');
    const titleEl = document.getElementById('modal-target-title');
    const subtitleEl = document.getElementById('modal-target-subtitle');
    if (titleEl) titleEl.textContent = targetName;
    if (subtitleEl) {
      subtitleEl.innerHTML = `${type === 'interface' ? 'Interface' : 'Simple Queue'} &bull; ${subtitle} &bull; Status: ${isRunning ? '<span class="text-emerald-600 font-semibold">Running</span>' : '<span class="text-amber-600 font-semibold">Offline</span>'}`;
    }

    this.trafficHistory = [];
    for (let i = 0; i < this.historyLength; i++) {
      this.trafficHistory.push({ rx: 0, tx: 0 });
    }
    this.peakRx = 0;
    this.peakTx = 0;

    ['rx-rate-display', 'tx-rate-display'].forEach(id => {
      const el = document.getElementById(id);
      if (el) el.textContent = '0 bps';
    });
    ['rx-peak-display', 'tx-peak-display'].forEach(id => {
      const el = document.getElementById(id);
      if (el) el.textContent = 'Peak: 0 bps';
    });
    const totalBytesEl = document.getElementById('total-bytes-display');
    const packetsEl = document.getElementById('packets-display');
    if (totalBytesEl) {
      totalBytesEl.textContent = '-';
      totalBytesEl.removeAttribute('title');
    }
    if (packetsEl) {
      packetsEl.textContent = '-';
      packetsEl.removeAttribute('title');
    }

    if (modal) modal.classList.remove('hidden');
    this.trafficController = new AbortController();
    this.drawChart();

    if (type === 'interface') {
      this.startInterfaceTrafficStream(targetName);
    } else {
      this.startQueueTrafficStream(targetName);
    }
  },

  startInterfaceTrafficStream(ifaceName) {
    API.stream(
      '/polyglot.v1.DeviceService/StreamInterfaceTraffic',
      { id: this.currentDeviceId, interfaceName: ifaceName },
      (frame) => {
        const rx = Number(frame.rxBps || frame.rx_bps || 0);
        const tx = Number(frame.txBps || frame.tx_bps || 0);
        this.updateTrafficMetrics(rx, tx);
      },
      (err) => { console.warn('Interface stream closed:', err); },
      this.trafficController.signal
    );
  },

  startQueueTrafficStream(queueName) {
    API.stream(
      '/polyglot.v1.NetworkMonitorService/StreamQueueStats',
      { deviceId: this.currentDeviceId, name: queueName, interval: '1s' },
      (frame) => {
        const qList = frame.queues || [];
        const q = qList.find(item => item.name === queueName) || qList[0];
        if (!q) return;

        let rx = 0, tx = 0;
        if (q.rate && q.rate.includes('/')) {
          const parts = q.rate.split('/');
          rx = Number(parts[0]) || 0;
          tx = Number(parts[1]) || 0;
        }

        const totalBytesEl = document.getElementById('total-bytes-display');
        const packetsEl = document.getElementById('packets-display');
        if (totalBytesEl && q.bytes) {
          if (q.bytes.includes('/')) {
            const bParts = q.bytes.split('/');
            totalBytesEl.textContent = `${this.formatBytes(bParts[0])} / ${this.formatBytes(bParts[1])}`;
            totalBytesEl.setAttribute('title', `Upload: ${this.formatBytes(bParts[0])} (${bParts[0]} B) | Download: ${this.formatBytes(bParts[1])} (${bParts[1]} B)`);
          } else {
            totalBytesEl.textContent = this.formatBytes(q.bytes);
            totalBytesEl.setAttribute('title', `${q.bytes} B`);
          }
        }
        if (packetsEl && q.packets) {
          packetsEl.textContent = this.formatPackets(q.packets);
          const dropInfo = q.dropped ? ` | Dropped: ${this.formatPackets(q.dropped)}` : '';
          packetsEl.setAttribute('title', `Packets: ${q.packets}${q.dropped ? ' | Dropped: ' + q.dropped : ''}`);
        }

        this.updateTrafficMetrics(rx, tx);
      },
      (err) => { console.warn('Queue stream closed:', err); },
      this.trafficController.signal
    );
  },

  updateTrafficMetrics(rx, tx) {
    if (rx > this.peakRx) this.peakRx = rx;
    if (tx > this.peakTx) this.peakTx = tx;

    const rxEl = document.getElementById('rx-rate-display');
    const txEl = document.getElementById('tx-rate-display');
    const rxPeakEl = document.getElementById('rx-peak-display');
    const txPeakEl = document.getElementById('tx-peak-display');

    if (rxEl) rxEl.textContent = this.formatBps(rx);
    if (txEl) txEl.textContent = this.formatBps(tx);
    if (rxPeakEl) rxPeakEl.textContent = `Peak: ${this.formatBps(this.peakRx)}`;
    if (txPeakEl) txPeakEl.textContent = `Peak: ${this.formatBps(this.peakTx)}`;

    this.trafficHistory.push({ rx, tx });
    if (this.trafficHistory.length > this.historyLength) {
      this.trafficHistory.shift();
    }
    this.drawChart();
  },

  closeTrafficModal() {
    if (this.trafficController) {
      this.trafficController.abort();
      this.trafficController = null;
    }
    const modal = document.getElementById('traffic-modal');
    if (modal) modal.classList.add('hidden');
  },

  drawChart() {
    const canvas = document.getElementById('traffic-canvas');
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    const dpr = window.devicePixelRatio || 1;
    const rect = canvas.getBoundingClientRect();
    if (canvas.width !== rect.width * dpr || canvas.height !== rect.height * dpr) {
      canvas.width = rect.width * dpr;
      canvas.height = rect.height * dpr;
    }

    const { width, height } = canvas;
    ctx.clearRect(0, 0, width, height);

    const maxY = Math.max(this.peakRx, this.peakTx, 100000) * 1.15;
    ctx.strokeStyle = 'rgba(226, 232, 240, 0.9)';
    ctx.lineWidth = 1 * dpr;

    for (let i = 1; i <= 3; i++) {
      const y = height - (height * (i / 4));
      ctx.beginPath();
      ctx.moveTo(0, y);
      ctx.lineTo(width, y);
      ctx.stroke();

      ctx.fillStyle = 'rgba(100, 116, 139, 0.8)';
      ctx.font = `${9 * dpr}px monospace`;
      ctx.fillText(this.formatBps(maxY * (i / 4)), 8 * dpr, y - 4 * dpr);
    }

    const stepX = width / (this.historyLength - 1);
    const renderSeries = (prop, strokeColor, fillColor) => {
      ctx.beginPath();
      const fillPath = new Path2D();
      this.trafficHistory.forEach((pt, i) => {
        const x = i * stepX;
        const y = height - ((pt[prop] || 0) / maxY) * height;
        if (i === 0) {
          ctx.moveTo(x, y);
          fillPath.moveTo(x, y);
        } else {
          ctx.lineTo(x, y);
          fillPath.lineTo(x, y);
        }
      });
      fillPath.lineTo(width, height);
      fillPath.lineTo(0, height);
      fillPath.closePath();

      const gradient = ctx.createLinearGradient(0, 0, 0, height);
      gradient.addColorStop(0, fillColor);
      gradient.addColorStop(1, 'rgba(255, 255, 255, 0)');
      ctx.fillStyle = gradient;
      ctx.fill(fillPath);

      ctx.strokeStyle = strokeColor;
      ctx.lineWidth = 2 * dpr;
      ctx.stroke();
    };

    renderSeries('rx', '#059669', 'rgba(16, 185, 129, 0.20)');
    renderSeries('tx', '#4f46e5', 'rgba(99, 102, 241, 0.20)');
  },

  formatBps(bps) {
    if (!bps || bps <= 0) return '0 bps';
    if (bps >= 1000000000) return (bps / 1000000000).toFixed(2) + ' Gbps';
    if (bps >= 1000000) return (bps / 1000000).toFixed(2) + ' Mbps';
    if (bps >= 1000) return (bps / 1000).toFixed(1) + ' Kbps';
    return bps + ' bps';
  },

  formatCount(val) {
    const n = Number(val);
    if (isNaN(n) || n < 0) return val || '0';
    if (n >= 1e9) return (n / 1e9).toFixed(2) + 'B';
    if (n >= 1e6) return (n / 1e6).toFixed(2) + 'M';
    if (n >= 1e3) return (n / 1e3).toFixed(1) + 'k';
    return n.toLocaleString();
  },

  formatPackets(pktStr) {
    if (!pktStr || pktStr === '0') return '0 pkt';
    if (pktStr.includes('/')) {
      const parts = pktStr.split('/');
      return `${this.formatCount(parts[0])} / ${this.formatCount(parts[1])} pkt`;
    }
    return `${this.formatCount(pktStr)} pkt`;
  },

  formatBytes(bytes) {
    const b = Number(bytes);
    if (!b || isNaN(b) || b <= 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(b) / Math.log(k));
    return parseFloat((b / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  }
};

window.MonitorInterface = MonitorInterface;
