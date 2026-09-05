/**
 * Polyglot Web Dial MikroTik Logs Controller
 * Menghadirkan fitur selaras dengan Portal Utama (web/src/features/logs):
 * - Dual View: Structured Table vs Raw Terminal macOS
 * - Severity Pills (All, Error, Warning, Info), Topic Dropdown, Search
 * - Pause/Resume, Clear, Auto-scroll, Floating Jump to Latest Button
 * - Ekspor TXT & JSON
 * - ConnectRPC Streaming melalui application/connect+json
 */

const MikroTikLogs = {
  currentDeviceId: null,
  logs: [], // Riwayat entri log (maksimal 1000)
  maxLogs: 1000,
  seenFingerprints: new Set(),

  // Filter & State
  viewMode: localStorage.getItem('logs_view_mode') || 'table', // 'table' | 'raw'
  severityFilter: 'all', // 'all' | 'error' | 'warning' | 'info'
  selectedTopic: '',
  searchTerm: '',
  isPaused: false,
  isAutoScroll: true,

  // Stream Lifecycle
  streamController: null,
  isStreaming: false,

  init() {
    // Terapkan view mode tersimpan
    this.switchView(this.viewMode, false);

    // Hubungkan selector router
    RouterSelector.onDeviceChange((deviceId) => {
      if (!deviceId) return;
      this.currentDeviceId = deviceId;
      this.refresh();
    });

    const initialDevId = RouterSelector.getSelectedDeviceId();
    if (initialDevId && !this.currentDeviceId) {
      this.currentDeviceId = initialDevId;
      this.refresh();
    }

    // Inisialisasi search input
    const searchInput = document.getElementById('log-search-input');
    const clearSearchBtn = document.getElementById('clear-search-btn');
    if (searchInput) {
      searchInput.oninput = (e) => {
        this.searchTerm = (e.target.value || '').trim().toLowerCase();
        if (clearSearchBtn) clearSearchBtn.classList.toggle('hidden', !this.searchTerm);
        this.renderLogs();
      };
    }

    // Pasang scroll listener untuk tombol Jump to Latest
    const tableContainer = document.getElementById('container-table-view');
    const rawContainer = document.getElementById('container-raw-view');
    [tableContainer, rawContainer].forEach(el => {
      if (el) {
        el.addEventListener('scroll', () => this.handleScroll(el));
      }
    });

    // Clean up on page navigation
    window.addEventListener('beforeunload', () => {
      if (this.streamController) {
        this.streamController.abort();
        this.streamController = null;
      }
    });

    // Tutup export dropdown jika klik di luar
    document.addEventListener('click', (e) => {
      const exportMenu = document.getElementById('export-dropdown');
      const exportBtn = document.getElementById('export-menu-btn');
      if (exportMenu && !exportMenu.contains(e.target) && !exportBtn.contains(e.target)) {
        exportMenu.classList.add('hidden');
      }
    });
  },

  cleanupStream() {
    if (this.streamController) {
      try {
        this.streamController.abort();
      } catch (_) {}
      this.streamController = null;
    }
  },

  startStream() {
    if (!this.currentDeviceId) return;

    this.cleanupStream();

    const icon = document.getElementById('refresh-icon');
    if (icon) icon.classList.add('animate-spin');
    setTimeout(() => {
      if (icon) icon.classList.remove('animate-spin');
    }, 800);

    this.streamController = new AbortController();
    const signal = this.streamController.signal;

    // Follow streaming: Biarkan stream tetap aktif mengalirkan hingga 1000 log secara realtime
    API.stream(
      '/polyglot.v1.NetworkMonitorService/StreamLogs',
      {
        deviceId: this.currentDeviceId,
        topics: '' // Mengalirkan seluruh topik dari router sehingga filter topik instan di client
      },
      (frame) => {
        const incoming = frame.logs || [];
        if (incoming.length === 0) return;

        let added = false;
        incoming.forEach(item => {
          const fp = `${item.id || ''}|${item.time || ''}|${item.topics || ''}|${item.message || ''}`;
          if (!this.seenFingerprints.has(fp)) {
            this.seenFingerprints.add(fp);
            const severity = this.classifySeverity(item.topics || '', item.message || '');
            this.logs.push({
              id: item.id || String(Date.now()),
              time: item.time || '',
              topics: item.topics || 'system',
              message: item.message || '',
              severity
            });
            added = true;
          }
        });

        if (added) {
          if (this.logs.length > this.maxLogs) {
            const removed = this.logs.splice(0, this.logs.length - this.maxLogs);
            removed.forEach(r => {
              this.seenFingerprints.delete(`${r.id}|${r.time}|${r.topics}|${r.message}`);
            });
          }
          this.updateTopicDropdown();
          this.renderLogs();
        }
      },
      (err) => {
        if (signal.aborted) return;
        setTimeout(() => {
          if (!signal.aborted) this.startStream();
        }, 5000);
      },
      signal
    );
  },

  refresh() {
    this.logs = [];
    this.seenFingerprints.clear();
    this._lastTopicKey = '';
    this.updateTopicDropdown();
    this.renderLogs();
    this.startStream();
  },

  userScrolledUp: false,
  _lastTopicKey: '',

  updateTopicDropdown() {
    const select = document.getElementById('topic-select');
    if (!select || document.activeElement === select) return;

    const currentVal = this.selectedTopic;
    const topicCounts = new Map();

    this.logs.forEach(l => {
      if (!l.topics) return;
      const tokens = l.topics.split(',').map(t => t.trim()).filter(Boolean);
      tokens.forEach(t => {
        topicCounts.set(t, (topicCounts.get(t) || 0) + 1);
      });
    });

    const sortedTopics = Array.from(topicCounts.keys()).sort();
    const optionsKey = sortedTopics.map(t => `${t}:${topicCounts.get(t)}`).join('|');
    if (this._lastTopicKey === optionsKey) return;
    this._lastTopicKey = optionsKey;

    let html = `<option value="">Semua Topik (${this.logs.length})</option>`;
    sortedTopics.forEach(t => {
      const count = topicCounts.get(t);
      const sel = (t.toLowerCase() === currentVal.toLowerCase()) ? ' selected' : '';
      html += `<option value="${this.escapeHtml(t)}"${sel}>${this.escapeHtml(t)} (${count})</option>`;
    });

    select.innerHTML = html;
    select.value = currentVal;
  },

  classifySeverity(topics, message) {
    const t = (topics || '').toLowerCase();
    const m = (message || '').toLowerCase();

    if (
      t.includes('error') || t.includes('critical') ||
      m.includes('failure') || m.includes('failed') ||
      m.includes('denied') || m.includes('rejected') ||
      m.includes('unreachable') || m.includes('kernel error')
    ) {
      return 'error';
    }

    if (
      t.includes('warning') || t.includes('warn') ||
      m.includes('warn') || m.includes('limit reached') ||
      m.includes('exceeded') || m.includes('timeout') ||
      m.includes('disconnected')
    ) {
      return 'warning';
    }

    if (t.includes('debug')) return 'debug';
    return 'info';
  },

  getFilteredLogs() {
    return this.logs.filter(item => {
      if (this.severityFilter !== 'all') {
        if (this.severityFilter === 'info') {
          if (item.severity !== 'info' && item.severity !== 'debug') return false;
        } else if (item.severity !== this.severityFilter) {
          return false;
        }
      }
      if (this.selectedTopic) {
        const itemTopics = (item.topics || '').toLowerCase();
        const target = this.selectedTopic.toLowerCase();
        const topicList = itemTopics.split(',').map(s => s.trim());
        if (!topicList.includes(target) && !itemTopics.includes(target)) {
          return false;
        }
      }
      if (this.searchTerm) {
        const q = this.searchTerm;
        return (
          (item.message || '').toLowerCase().includes(q) ||
          (item.topics || '').toLowerCase().includes(q) ||
          (item.time || '').toLowerCase().includes(q)
        );
      }
      return true;
    });
  },

  renderLogs() {
    const filtered = this.getFilteredLogs();
    const counterEl = document.getElementById('log-counter');
    if (counterEl) {
      counterEl.textContent = `Menampilkan ${filtered.length} dari ${this.logs.length} log (Maks. ${this.maxLogs})`;
    }

    if (this.viewMode === 'table') {
      this.renderTable(filtered);
    } else {
      this.renderRaw(filtered);
    }

    if (this.isAutoScroll && !this.userScrolledUp) {
      requestAnimationFrame(() => {
        this.scrollToBottom(false);
      });
    }
  },

  renderTable(logs) {
    const tbody = document.getElementById('logs-table-tbody');
    if (!tbody) return;

    if (logs.length === 0) {
      tbody.innerHTML = `
        <tr>
          <td colspan="5" class="py-12 text-center text-slate-400 font-mono">
            ${this.logs.length === 0 ? 'Menunggu aliran log dari router...' : 'Tidak ada log yang cocok dengan filter.'}
          </td>
        </tr>
      `;
      return;
    }

    tbody.innerHTML = logs.map(l => {
      let sevBadge = '';
      if (l.severity === 'error') {
        sevBadge = '<span class="inline-flex items-center px-2 py-0.5 rounded text-[10px] font-bold uppercase bg-rose-50 text-rose-700 border border-rose-200">ERROR</span>';
      } else if (l.severity === 'warning') {
        sevBadge = '<span class="inline-flex items-center px-2 py-0.5 rounded text-[10px] font-bold uppercase bg-amber-50 text-amber-700 border border-amber-200">WARN</span>';
      } else {
        sevBadge = '<span class="inline-flex items-center px-2 py-0.5 rounded text-[10px] font-semibold uppercase bg-blue-50 text-blue-700 border border-blue-200">INFO</span>';
      }

      const highlightedMsg = this.highlightText(l.message, this.searchTerm);

      const topicBadges = (l.topics || 'system').split(',').map(t => {
        const trimmed = t.trim();
        const isSelected = this.selectedTopic.toLowerCase() === trimmed.toLowerCase();
        const badgeClass = isSelected
          ? 'bg-blue-600 text-white font-bold shadow-xs'
          : 'bg-slate-100 text-slate-600 hover:bg-blue-50 hover:text-blue-700';
        return `<button onclick="MikroTikLogs.selectTopicBadge('${this.escapeHtml(trimmed)}')" title="Filter topik ${this.escapeHtml(trimmed)}" class="inline-block px-1.5 py-0.5 rounded text-[10px] font-mono cursor-pointer transition-colors ${badgeClass}">${this.escapeHtml(trimmed)}</button>`;
      }).join(' ');

      return `
        <tr class="hover:bg-slate-50/80 transition-colors font-mono">
          <td class="py-2.5 px-4 text-[11px] text-slate-500 whitespace-nowrap">${l.time || '-'}</td>
          <td class="py-2.5 px-4 whitespace-nowrap">${sevBadge}</td>
          <td class="py-2.5 px-4 whitespace-nowrap">
            <div class="flex flex-wrap items-center gap-1">${topicBadges}</div>
          </td>
          <td class="py-2.5 px-4 text-[11px] text-slate-800 break-words leading-relaxed">${highlightedMsg}</td>
          <td class="py-2.5 px-4 text-right">
            <button onclick="MikroTikLogs.copyLog('${encodeURIComponent(l.message)}')" title="Salin Pesan" class="p-1 text-slate-400 hover:text-slate-700 rounded transition-colors cursor-pointer">
              <svg xmlns="http://www.w3.org/2000/svg" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect width="14" height="14" x="8" y="8" rx="2" ry="2"/><path d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2"/></svg>
            </button>
          </td>
        </tr>
      `;
    }).join('');
  },

  renderRaw(logs) {
    const container = document.getElementById('raw-logs-content');
    if (!container) return;

    if (logs.length === 0) {
      container.innerHTML = `<div class="text-slate-500 text-center py-12">${this.logs.length === 0 ? 'Menghubungkan ke stream log router...' : 'Tidak ada log yang cocok dengan filter.'}</div>`;
      return;
    }

    container.innerHTML = logs.map(l => {
      let sevColor = 'text-blue-400';
      if (l.severity === 'error') sevColor = 'text-rose-400 font-bold';
      else if (l.severity === 'warning') sevColor = 'text-amber-400 font-bold';

      const highlightedMsg = this.highlightText(l.message, this.searchTerm);

      const rawTopicBadges = (l.topics || 'system').split(',').map(t => {
        const trimmed = t.trim();
        const isSelected = this.selectedTopic.toLowerCase() === trimmed.toLowerCase();
        const badgeClass = isSelected
          ? 'bg-blue-600 text-white'
          : 'bg-slate-900 text-slate-400 border border-slate-800 hover:text-white';
        return `<button onclick="MikroTikLogs.selectTopicBadge('${this.escapeHtml(trimmed)}')" title="Filter topik ${this.escapeHtml(trimmed)}" class="text-[10px] px-1 rounded cursor-pointer ${badgeClass}">${this.escapeHtml(trimmed)}</button>`;
      }).join(' ');

      return `
        <div class="hover:bg-slate-900/80 px-2 py-0.5 rounded flex items-start gap-2.5 leading-relaxed">
          <span class="text-slate-500 text-[11px] shrink-0 select-none">${l.time}</span>
          <span class="${sevColor} text-[11px] shrink-0 uppercase w-14 font-semibold">[${l.severity}]</span>
          <div class="flex items-center gap-1 shrink-0">${rawTopicBadges}</div>
          <span class="text-slate-200 text-[11px] break-words flex-1">${highlightedMsg}</span>
        </div>
      `;
    }).join('');
  },

  highlightText(text, query) {
    if (!query) return this.escapeHtml(text);
    const escapedText = this.escapeHtml(text);
    const escapedQuery = this.escapeHtml(query);
    const regex = new RegExp(`(${escapedQuery})`, 'gi');
    return escapedText.replace(regex, '<mark class="bg-yellow-300 text-slate-900 px-0.5 rounded">$1</mark>');
  },

  escapeHtml(str) {
    return (str || '')
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#039;');
  },

  switchView(mode, render = true) {
    this.viewMode = mode;
    localStorage.setItem('logs_view_mode', mode);

    const tableBtn = document.getElementById('view-table-btn');
    const rawBtn = document.getElementById('view-raw-btn');
    const tableContainer = document.getElementById('container-table-view');
    const rawContainer = document.getElementById('container-raw-view');

    const activeCls = 'px-2.5 py-1 rounded-md text-[11px] font-bold bg-white text-slate-900 shadow-xs cursor-pointer transition-all flex items-center gap-1';
    const inactiveCls = 'px-2.5 py-1 rounded-md text-[11px] font-semibold text-slate-600 hover:text-slate-900 cursor-pointer transition-all flex items-center gap-1';

    if (mode === 'table') {
      if (tableBtn) tableBtn.className = activeCls;
      if (rawBtn) rawBtn.className = inactiveCls;
      if (tableContainer) tableContainer.classList.remove('hidden');
      if (rawContainer) rawContainer.classList.add('hidden');
    } else {
      if (tableBtn) tableBtn.className = inactiveCls;
      if (rawBtn) rawBtn.className = activeCls;
      if (tableContainer) tableContainer.classList.add('hidden');
      if (rawContainer) rawContainer.classList.remove('hidden');
    }

    if (render) this.renderLogs();
  },

  setSeverity(sev) {
    this.severityFilter = sev;
    ['all', 'error', 'warning', 'info'].forEach(s => {
      const btn = document.getElementById(`sev-${s}-btn`);
      if (btn) {
        if (s === sev) {
          btn.className = 'px-2.5 py-1 rounded-lg font-bold bg-white text-slate-900 shadow-xs cursor-pointer transition-all';
        } else {
          let textCls = 'text-slate-600';
          if (s === 'error') textCls = 'text-rose-600';
          if (s === 'warning') textCls = 'text-amber-600';
          if (s === 'info') textCls = 'text-blue-600';
          btn.className = `px-2.5 py-1 rounded-lg font-semibold ${textCls} hover:bg-slate-200/50 cursor-pointer transition-all`;
        }
      }
    });
    this.renderLogs();
  },

  setTopic(topic) {
    this.selectedTopic = (topic || '').trim();
    const select = document.getElementById('topic-select');
    if (select && select.value !== this.selectedTopic) {
      select.value = this.selectedTopic;
    }
    this.renderLogs();
  },

  selectTopicBadge(topic) {
    if (this.selectedTopic.toLowerCase() === topic.toLowerCase()) {
      this.setTopic('');
    } else {
      this.setTopic(topic);
    }
  },

  clearSearch() {
    const input = document.getElementById('log-search-input');
    const clearBtn = document.getElementById('clear-search-btn');
    if (input) input.value = '';
    if (clearBtn) clearBtn.classList.add('hidden');
    this.searchTerm = '';
    this.renderLogs();
  },

  togglePause() {
    this.isPaused = !this.isPaused;
    const label = document.getElementById('pause-label');
    const icon = document.getElementById('pause-icon');
    if (label) label.textContent = this.isPaused ? 'Lanjutkan' : 'Jeda';
    if (icon) {
      icon.innerHTML = this.isPaused
        ? '<polygon points="5 3 19 12 5 21 5 3"/>'
        : '<rect x="6" y="4" width="4" height="16"/><rect x="14" y="4" width="4" height="16"/>';
    }
  },

  toggleAutoScroll(enabled) {
    this.isAutoScroll = enabled;
    if (enabled) {
      this.scrollToBottom(false);
    }
  },

  clearLogs() {
    this.logs = [];
    this.seenFingerprints.clear();
    this._lastTopicKey = '';
    this.updateTopicDropdown();
    this.renderLogs();
  },

  handleScroll(container) {
    const distanceToBottom = container.scrollHeight - container.scrollTop - container.clientHeight;
    this.userScrolledUp = distanceToBottom > 60;
    const jumpBtn = document.getElementById('jump-to-latest-btn');
    if (jumpBtn) {
      jumpBtn.classList.toggle('hidden', !this.userScrolledUp);
    }
  },

  scrollToBottom(smooth = true) {
    this.userScrolledUp = false;
    const jumpBtn = document.getElementById('jump-to-latest-btn');
    if (jumpBtn) jumpBtn.classList.add('hidden');

    const activeEl = this.viewMode === 'table'
      ? document.getElementById('container-table-view')
      : document.getElementById('container-raw-view');

    if (activeEl) {
      if (smooth) {
        activeEl.scrollTo({ top: activeEl.scrollHeight, behavior: 'smooth' });
      } else {
        activeEl.scrollTop = activeEl.scrollHeight;
      }
    }
  },

  toggleExportMenu() {
    const el = document.getElementById('export-dropdown');
    if (el) el.classList.toggle('hidden');
  },

  exportTxt() {
    const filtered = this.getFilteredLogs();
    const content = filtered
      .map(l => `[${l.time}] [${l.severity.toUpperCase()}] [${l.topics}] ${l.message}`)
      .join('\n');
    this.downloadFile(content, 'text/plain;charset=utf-8', `mikrotik-logs-${Date.now()}.txt`);
    this.toggleExportMenu();
  },

  exportJson() {
    const filtered = this.getFilteredLogs();
    const content = JSON.stringify(filtered, null, 2);
    this.downloadFile(content, 'application/json;charset=utf-8', `mikrotik-logs-${Date.now()}.json`);
    this.toggleExportMenu();
  },

  downloadFile(content, mimeType, filename) {
    const blob = new Blob([content], { type: mimeType });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
    showToast(`File ${filename} berhasil diunduh!`, 'success');
  },

  copyLog(encodedMsg) {
    const msg = decodeURIComponent(encodedMsg);
    navigator.clipboard.writeText(msg).then(() => {
      showToast('Pesan log disalin ke clipboard!', 'success');
    }).catch(() => {
      showToast('Gagal menyalin log', 'error');
    });
  },

  updateStreamBadge(status) {
    const badge = document.getElementById('stream-badge');
    if (!badge) return;

    if (status === 'live') {
      badge.className = 'inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-emerald-50 text-emerald-700 border border-emerald-200';
      badge.innerHTML = '<span class="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse"></span> Streaming Live';
    } else if (status === 'connecting') {
      badge.className = 'inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-blue-50 text-blue-700 border border-blue-200';
      badge.innerHTML = '<span class="w-1.5 h-1.5 rounded-full bg-blue-500 animate-spin"></span> Menghubungkan...';
    } else {
      badge.className = 'inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-slate-100 text-slate-600';
      badge.innerHTML = '<span class="w-1.5 h-1.5 rounded-full bg-slate-400"></span> Terputus';
    }
  }
};

window.MikroTikLogs = MikroTikLogs;
