/**
 * Polyglot Web Dial Settings Controller
 *
 * Two permission-gated sections:
 *   - users   (#users)   : technician accounts, requires user:read; writes require user:manage.
 *   - devices (#devices) : router access matrix, requires user:manage + device:read.
 *
 * The page itself is guarded by Auth.requireMenu('settings.html') in the HTML
 * bootstrap; this controller only decides which tabs and actions are usable.
 */

const Settings = {
  users: [],
  technicians: [],
  devices: [],
  activeTab: 'users',
  canReadUsers: false,
  canManageUsers: false,
  canManageDevices: false,
  editingUser: null,

  async init() {
    this.canReadUsers = Auth.canMenu('settings-users');
    this.canManageUsers = Auth.canAction('settings-users', 'manage');
    this.canManageDevices = Auth.canMenu('settings-devices');

    if (!this.canReadUsers && !this.canManageDevices) {
      window.location.replace(Auth.firstAccessibleRoute() || 'no-access.html');
      return;
    }

    const addBtn = document.getElementById('open-add-tech-btn');
    if (addBtn) addBtn.classList.toggle('hidden', !this.canManageUsers);

    const readonlyNote = document.getElementById('users-readonly-note');
    if (readonlyNote) readonlyNote.classList.toggle('hidden', this.canManageUsers);

    this.bindTabs();
    this.bindModals();

    await this.loadDevices();
    await this.loadUsers();

    const initialTab = this.resolveTab((window.location.hash || '').replace('#', ''));
    this.switchTab(initialTab, { updateHash: false });
    this.renderTable();
    this.renderMatrix();
  },

  resolveTab(tab) {
    if (tab === 'devices' && this.canManageDevices) return 'devices';
    if (tab === 'users' && this.canReadUsers) return 'users';
    return this.canReadUsers ? 'users' : 'devices';
  },

  bindTabs() {
    const usersBtn = document.getElementById('settings-tab-users');
    const devicesBtn = document.getElementById('settings-tab-devices');
    const tabsEl = document.getElementById('settings-tabs');

    if (usersBtn) usersBtn.classList.toggle('hidden', !this.canReadUsers);
    if (devicesBtn) devicesBtn.classList.toggle('hidden', !this.canManageDevices);
    if (tabsEl) {
      const visible = [this.canReadUsers, this.canManageDevices].filter(Boolean).length;
      tabsEl.classList.toggle('hidden', visible < 2);
    }

    document.querySelectorAll('[data-settings-tab]').forEach((btn) => {
      btn.onclick = () => this.switchTab(btn.getAttribute('data-settings-tab'));
    });

    window.addEventListener('hashchange', () => {
      this.switchTab(this.resolveTab((window.location.hash || '').replace('#', '')), { updateHash: false });
    });

    window.addEventListener('languagechange', () => {
      this.renderTable();
      this.renderMatrix();
    });
  },

  switchTab(tab, options = {}) {
    const next = this.resolveTab(tab);
    this.activeTab = next;

    const usersPanel = document.getElementById('settings-panel-users');
    const devicesPanel = document.getElementById('settings-panel-devices');
    if (usersPanel) usersPanel.classList.toggle('hidden', next !== 'users');
    if (devicesPanel) devicesPanel.classList.toggle('hidden', next !== 'devices');

    document.querySelectorAll('[data-settings-tab]').forEach((btn) => {
      const isActive = btn.getAttribute('data-settings-tab') === next;
      btn.classList.toggle('bg-white', isActive);
      btn.classList.toggle('text-slate-900', isActive);
      btn.classList.toggle('shadow-xs', isActive);
      btn.classList.toggle('text-slate-500', !isActive);
      btn.classList.toggle('hover:text-slate-800', !isActive);
      btn.setAttribute('aria-selected', String(isActive));
    });

    if (next === 'devices') this.renderMatrix();

    if (options.updateHash !== false) {
      history.replaceState(null, '', `${window.location.pathname}${window.location.search}#${next}`);
    }
  },

  async loadDevices() {
    if (!Auth.can('dial:settings:devices:read') && !Auth.can('device:read')) {
      this.devices = [];
      return;
    }

    try {
      const res = await API.call('/polyglot.v1.DeviceService/ListDevices', {});
      this.devices = res.devices || [];
    } catch (err) {
      console.error('Failed to load admin devices:', err);
      this.devices = [];
      API.reportError(err, 'Gagal memuat daftar router');
    }
  },

  async loadUsers() {
    const loadingEl = document.getElementById('loading-state');
    const tableEl = document.getElementById('tech-table');
    const emptyEl = document.getElementById('empty-state');

    if (!Auth.can('dial:settings:users:read') && !Auth.can('user:read')) {
      this.users = [];
      this.technicians = [];
      if (loadingEl) loadingEl.classList.add('hidden');
      if (emptyEl) emptyEl.classList.remove('hidden');
      return;
    }

    if (loadingEl) loadingEl.classList.remove('hidden');
    if (tableEl) tableEl.classList.add('hidden');
    if (emptyEl) emptyEl.classList.add('hidden');

    try {
      const res = await API.call('/polyglot.v1.UserService/ListUsers', {
        page: 1,
        pageSize: 100
      });

      this.users = res.users || [];
      this.technicians = this.users.filter((u) => {
        const roles = (u.roles && u.roles.length > 0) ? u.roles : [u.role];
        return roles.some((r) => ['teknisi', 'agent'].includes((r || '').toLowerCase()));
      });
    } catch (err) {
      console.error('Failed to load technicians:', err);
      this.technicians = [];
      API.reportError(err, 'Gagal memuat data teknisi');
    } finally {
      if (loadingEl) loadingEl.classList.add('hidden');
      this.renderTable();
    }
  },

  renderTable() {
    const tbody = document.getElementById('tech-tbody');
    const tableEl = document.getElementById('tech-table');
    const emptyEl = document.getElementById('empty-state');
    if (!tbody) return;

    if (this.technicians.length === 0) {
      if (tableEl) tableEl.classList.add('hidden');
      if (emptyEl) emptyEl.classList.remove('hidden');
      return;
    }

    if (tableEl) tableEl.classList.remove('hidden');
    if (emptyEl) emptyEl.classList.add('hidden');

    const deviceMap = new Map(this.devices.map((d) => [String(d.id), d.name || d.host]));
    const canAssign = Auth.canAction('settings-devices', 'assign');

    tbody.innerHTML = this.technicians.map((tech, idx) => {
      const assignedIds = tech.assignedDeviceIds || tech.assigned_device_ids || [];
      const assignedBadges = assignedIds.length > 0
        ? assignedIds.map((id) => {
            const devName = deviceMap.get(String(id)) || id;
            return `<span class="inline-block px-2 py-0.5 rounded text-[11px] font-medium bg-blue-50 text-blue-700 border border-blue-200 mr-1 mb-1">${this.esc(devName)}</span>`;
          }).join('')
        : '<span class="text-xs text-slate-400 italic">Belum ada router di-assign</span>';

      const isActive = tech.isActive !== undefined ? tech.isActive : tech.is_active;
      const techId = String(tech.id);
      const fullName = tech.fullName || tech.full_name || tech.username;
      const phone = tech.phoneNumber || tech.phone_number || '';

      const actions = [];
      if (canAssign) {
        actions.push(`
          <button onclick="Settings.openAssignModal('${this.esc(techId)}')" class="px-2.5 py-1 text-xs bg-blue-50 hover:bg-blue-100 text-blue-700 rounded-lg font-semibold transition-colors cursor-pointer">
            Pilih Router
          </button>
        `);
      }
      if (this.canManageUsers) {
        actions.push(`
          <button onclick="Settings.toggleActive('${this.esc(techId)}', ${isActive})" class="px-2 py-1 text-xs border border-slate-200 hover:bg-slate-100 text-slate-700 rounded-lg font-medium transition-colors cursor-pointer">
            ${isActive ? 'Nonaktifkan' : 'Aktifkan'}
          </button>
        `);
      }

      return `
        <tr class="hover:bg-slate-50 transition-colors border-b border-slate-100 text-sm">
          <td class="py-3 px-4 font-mono text-xs text-slate-400">${idx + 1}</td>
          <td class="py-3 px-4">
            <div class="font-bold text-slate-900">${this.esc(fullName)}</div>
            <div class="text-xs text-slate-500 font-mono">@${this.esc(tech.username)}</div>
          </td>
          <td class="py-3 px-4 text-xs text-slate-600">
            <div>${this.esc(tech.email || '-')}</div>
            <div class="text-slate-400">${this.esc(phone)}</div>
          </td>
          <td class="py-3 px-4 max-w-xs">
            ${assignedBadges}
          </td>
          <td class="py-3 px-4">
            <span class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-semibold ${
              isActive ? 'bg-emerald-50 text-emerald-700' : 'bg-slate-100 text-slate-600'
            }">
              ${isActive ? 'Aktif' : 'Non-Aktif'}
            </span>
          </td>
          <td class="py-3 px-4 text-right">
            <div class="flex items-center justify-end gap-1.5">
              ${actions.length > 0 ? actions.join('') : '<span class="text-xs text-slate-400">-</span>'}
            </div>
          </td>
        </tr>
      `;
    }).join('');
  },

  /**
   * Renders the user x router access matrix. Each checkbox calls
   * AssignDevicesToUser with the row's complete selection, so a single toggle
   * is atomic and does not depend on other rows.
   */
  renderMatrix() {
    const container = document.getElementById('access-matrix-container');
    const loadingEl = document.getElementById('matrix-loading');
    const emptyEl = document.getElementById('matrix-empty');
    if (!container) return;

    if (loadingEl) loadingEl.classList.add('hidden');

    if (!this.canManageDevices || this.technicians.length === 0 || this.devices.length === 0) {
      container.classList.add('hidden');
      if (emptyEl) emptyEl.classList.remove('hidden');
      return;
    }

    if (emptyEl) emptyEl.classList.add('hidden');
    container.classList.remove('hidden');

    const header = this.devices.map((d) => `
      <th class="py-3 px-3 min-w-[170px] border-l border-slate-200/70 text-left">
        <div class="font-bold text-slate-700 truncate max-w-[190px]" title="${this.esc(d.name || d.host)}">${this.esc(d.name || d.host)}</div>
        <div class="text-[10px] font-mono text-slate-400 truncate max-w-[190px]">${this.esc(d.host)}${d.port ? ':' + this.esc(d.port) : ''}</div>
      </th>
    `).join('');

    const rows = this.technicians.map((tech) => {
      const assigned = new Set((tech.assignedDeviceIds || tech.assigned_device_ids || []).map(String));
      const cells = this.devices.map((d) => `
        <td class="py-3 px-3 border-l border-slate-100 text-center">
          <input type="checkbox"
            data-access-checkbox
            data-user-id="${this.esc(tech.id)}"
            data-device-id="${this.esc(d.id)}"
            ${assigned.has(String(d.id)) ? 'checked' : ''}
            class="w-4 h-4 rounded border-slate-300 text-blue-600 focus:ring-blue-500 cursor-pointer">
        </td>
      `).join('');

      const fullName = tech.fullName || tech.full_name || tech.username;
      return `
        <tr class="hover:bg-slate-50/70 transition-colors border-b border-slate-100">
          <td class="py-3 px-4">
            <div class="font-bold text-slate-900 text-sm">${this.esc(fullName)}</div>
            <div class="text-[10px] text-slate-500 font-mono">@${this.esc(tech.username)}</div>
          </td>
          ${cells}
        </tr>
      `;
    }).join('');

    container.innerHTML = `
      <table class="w-full text-left border-collapse">
        <thead>
          <tr class="bg-slate-50/80 border-b border-slate-200 text-xs font-semibold uppercase tracking-wider text-slate-600">
            <th class="py-3 px-4 min-w-[200px]">Teknisi</th>
            ${header}
          </tr>
        </thead>
        <tbody>${rows}</tbody>
      </table>
    `;

    container.onchange = (e) => {
      const checkbox = e.target.closest('[data-access-checkbox]');
      if (checkbox) this.handleMatrixToggle(checkbox);
    };
  },

  async handleMatrixToggle(checkbox) {
    if (!this.canManageDevices) {
      checkbox.checked = !checkbox.checked;
      return;
    }

    const userId = checkbox.getAttribute('data-user-id');
    const row = checkbox.closest('tr');
    const checkboxes = row ? Array.from(row.querySelectorAll('[data-access-checkbox]')) : [checkbox];
    const deviceIds = checkboxes
      .filter((cb) => cb.checked)
      .map((cb) => cb.getAttribute('data-device-id'));

    checkboxes.forEach((cb) => { cb.disabled = true; });

    try {
      await API.call('/polyglot.v1.UserService/AssignDevicesToUser', {
        userId: Number(userId),
        deviceIds
      });

      const tech = this.technicians.find((u) => String(u.id) === String(userId));
      if (tech) {
        tech.assignedDeviceIds = deviceIds;
        tech.assigned_device_ids = deviceIds;
      }
      showToast('Akses router berhasil diperbarui', 'success');
      this.renderTable();
    } catch (err) {
      checkbox.checked = !checkbox.checked;
      API.reportError(err, 'Gagal memperbarui akses router');
    } finally {
      checkboxes.forEach((cb) => { cb.disabled = false; });
    }
  },

  bindModals() {
    const addBtn = document.getElementById('open-add-tech-btn');
    const addModal = document.getElementById('add-tech-modal');
    const closeAddBtn = document.getElementById('close-add-modal-btn');
    const addForm = document.getElementById('add-tech-form');

    if (addBtn && addModal) {
      addBtn.onclick = () => {
        if (!this.canManageUsers) return;
        this.renderRouterCheckboxes('add-router-checkboxes');
        addModal.classList.remove('hidden');
      };
    }
    if (closeAddBtn && addModal) {
      closeAddBtn.onclick = () => addModal.classList.add('hidden');
    }
    if (addModal) {
      addModal.onclick = (e) => {
        if (e.target === addModal) addModal.classList.add('hidden');
      };
    }

    if (addForm) {
      addForm.onsubmit = async (e) => {
        e.preventDefault();
        if (!this.canManageUsers) return;

        const username = document.getElementById('tech-username').value.trim();
        const email = document.getElementById('tech-email').value.trim();
        const password = document.getElementById('tech-password').value;
        const fullName = document.getElementById('tech-fullname').value.trim();
        const phoneNumber = document.getElementById('tech-phone').value.trim();

        const selectedDevices = Array.from(
          document.querySelectorAll('#add-router-checkboxes input:checked')
        ).map((cb) => cb.value);

        if (selectedDevices.length === 0) {
          if (!confirm('Anda belum memilih router untuk teknisi ini. Lanjutkan pembuatan teknisi tanpa router?')) {
            return;
          }
        }

        const submitBtn = document.getElementById('submit-add-btn');
        if (submitBtn) submitBtn.disabled = true;

        try {
          await API.call('/polyglot.v1.UserService/CreateUser', {
            username,
            email,
            password,
            role: 'teknisi',
            fullName,
            phoneNumber,
            assignedDeviceIds: selectedDevices
          });

          showToast('Akun teknisi berhasil dibuat!', 'success');
          addModal.classList.add('hidden');
          addForm.reset();
          await this.loadUsers();
          this.renderMatrix();
        } catch (err) {
          API.reportError(err, 'Gagal membuat teknisi');
        } finally {
          if (submitBtn) submitBtn.disabled = false;
        }
      };
    }

    const assignModal = document.getElementById('assign-modal');
    const closeAssignBtn = document.getElementById('close-assign-modal-btn');
    const saveAssignBtn = document.getElementById('save-assign-btn');

    if (closeAssignBtn && assignModal) {
      closeAssignBtn.onclick = () => assignModal.classList.add('hidden');
    }
    if (assignModal) {
      assignModal.onclick = (e) => {
        if (e.target === assignModal) assignModal.classList.add('hidden');
      };
    }

    if (saveAssignBtn) {
      saveAssignBtn.onclick = async () => {
        if (!this.editingUser || !Auth.canAction('settings-devices', 'assign')) return;

        const selectedDevices = Array.from(
          document.querySelectorAll('#edit-router-checkboxes input:checked')
        ).map((cb) => cb.value);

        saveAssignBtn.disabled = true;

        try {
          await API.call('/polyglot.v1.UserService/AssignDevicesToUser', {
            userId: Number(this.editingUser.id),
            deviceIds: selectedDevices
          });

          showToast('Akses router teknisi berhasil diperbarui!', 'success');
          assignModal.classList.add('hidden');
          await this.loadUsers();
          this.renderMatrix();
        } catch (err) {
          API.reportError(err, 'Gagal memperbarui router teknisi');
        } finally {
          saveAssignBtn.disabled = false;
        }
      };
    }

    window.addEventListener('keydown', (e) => {
      if (e.key === 'Escape') {
        if (addModal) addModal.classList.add('hidden');
        if (assignModal) assignModal.classList.add('hidden');
      }
    });
  },

  renderRouterCheckboxes(containerId, preselectedIds = []) {
    const container = document.getElementById(containerId);
    if (!container) return;

    if (!this.devices || this.devices.length === 0) {
      container.innerHTML = '<p class="text-xs text-slate-400 italic p-3 text-center">Belum ada router di inventory sistem atau Anda tidak memiliki izin device:read.</p>';
      return;
    }

    const preselectedSet = new Set((preselectedIds || []).map(String));

    container.innerHTML = this.devices.map((d) => {
      const isChecked = preselectedSet.has(String(d.id));
      return `
        <label class="flex items-center gap-2.5 p-2.5 rounded-xl border border-slate-200 hover:bg-slate-100/80 cursor-pointer text-xs font-medium text-slate-800 transition-colors">
          <input type="checkbox" value="${this.esc(d.id)}" ${isChecked ? 'checked' : ''} class="w-4 h-4 rounded border-slate-300 text-blue-600 focus:ring-blue-500 cursor-pointer">
          <div class="flex flex-col">
            <span class="font-bold text-slate-900">${this.esc(d.name || d.host)}</span>
            <span class="text-[10px] text-slate-500 font-mono">${this.esc(d.host)}${d.port ? ':' + this.esc(d.port) : ''}</span>
          </div>
        </label>
      `;
    }).join('');
  },

  openAssignModal(techId) {
    if (!Auth.canAction('settings-devices', 'assign')) return;

    const tech = this.technicians.find((t) => String(t.id) === String(techId));
    if (!tech) return;

    this.editingUser = tech;
    const modal = document.getElementById('assign-modal');
    const targetName = document.getElementById('assign-target-name');

    if (targetName) targetName.textContent = tech.fullName || tech.full_name || tech.username;
    this.renderRouterCheckboxes('edit-router-checkboxes', tech.assignedDeviceIds || tech.assigned_device_ids || []);

    if (modal) modal.classList.remove('hidden');
  },

  async toggleActive(techId, currentActive) {
    if (!this.canManageUsers) return;
    try {
      await API.call('/polyglot.v1.UserService/ToggleActive', {
        id: Number(techId)
      });
      showToast('Status teknisi berhasil diubah', 'success');
      await this.loadUsers();
      this.renderMatrix();
    } catch (err) {
      API.reportError(err, 'Gagal mengubah status');
    }
  },

  esc(value) {
    return String(value === null || value === undefined ? '' : value)
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#039;');
  }
};

window.Settings = Settings;
