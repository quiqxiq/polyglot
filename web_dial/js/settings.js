/**
 * Polyglot Web Dial Settings & Technician Management Controller
 * KHUSUS UNTUK ROLE ADMIN:
 * Mengelola akun teknisi dan memilihkan router-router yang boleh diakses oleh teknisi tersebut.
 */

const Settings = {
  adminDevices: [],
  technicians: [],
  editingUser: null,

  async init() {
    // Pastikan hanya role Admin yang bisa menjalankan modul ini
    if (!Auth.isAdmin()) {
      alert('Akses Ditolak: Halaman Pengaturan hanya dapat diakses oleh Admin.');
      window.location.href = 'index.html';
      return;
    }

    this.bindModals();
    await this.loadAdminDevices();
    await this.loadTechnicians();
  },

  async loadAdminDevices() {
    try {
      // ListDevices untuk Admin otomatis hanya mengembalikan router milik Admin tersebut
      const res = await API.call('/polyglot.v1.DeviceService/ListDevices', {});
      this.adminDevices = res.devices || [];
    } catch (err) {
      console.error('Failed to load admin devices:', err);
      showToast('Gagal memuat daftar router Admin', 'error');
    }
  },

  async loadTechnicians() {
    const loadingEl = document.getElementById('loading-state');
    const tableEl = document.getElementById('tech-table');
    const emptyEl = document.getElementById('empty-state');

    if (loadingEl) loadingEl.classList.remove('hidden');
    if (tableEl) tableEl.classList.add('hidden');
    if (emptyEl) emptyEl.classList.add('hidden');

    try {
      const res = await API.call('/polyglot.v1.UserService/ListUsers', {
        page: 1,
        pageSize: 100
      });

      const allUsers = res.users || [];
      // Tampilkan akun dengan role teknisi / agent
      this.technicians = allUsers.filter(u => {
        const roles = (u.roles && u.roles.length > 0) ? u.roles : [u.role];
        return roles.some(r => ['teknisi', 'technician', 'agent'].includes((r || '').toLowerCase()));
      });

      this.renderTable();
    } catch (err) {
      console.error('Failed to load technicians:', err);
      showToast('Gagal memuat data teknisi: ' + err.message, 'error');
    } finally {
      if (loadingEl) loadingEl.classList.add('hidden');
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

    const deviceMap = new Map(this.adminDevices.map(d => [String(d.id), d.name || d.host]));

    tbody.innerHTML = this.technicians.map((tech, idx) => {
      const assignedIds = tech.assignedDeviceIds || tech.assigned_device_ids || [];
      const assignedBadges = assignedIds.length > 0
        ? assignedIds.map(id => {
            const devName = deviceMap.get(String(id)) || id;
            return `<span class="inline-block px-2 py-0.5 rounded text-[11px] font-medium bg-blue-50 text-blue-700 border border-blue-200 mr-1 mb-1">${devName}</span>`;
          }).join('')
        : '<span class="text-xs text-slate-400 italic">Belum ada router di-assign</span>';

      const isActive = tech.isActive !== undefined ? tech.isActive : tech.is_active;
      const techId = String(tech.id);
      const fullName = tech.fullName || tech.full_name || tech.username;
      const phone = tech.phoneNumber || tech.phone_number || '';

      return `
        <tr class="hover:bg-slate-50 transition-colors border-b border-slate-100 text-sm">
          <td class="py-3 px-4 font-mono text-xs text-slate-400">${idx + 1}</td>
          <td class="py-3 px-4">
            <div class="font-bold text-slate-900">${fullName}</div>
            <div class="text-xs text-slate-500 font-mono">@${tech.username}</div>
          </td>
          <td class="py-3 px-4 text-xs text-slate-600">
            <div>${tech.email || '-'}</div>
            <div class="text-slate-400">${phone}</div>
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
              <button onclick="Settings.openAssignModal('${techId}')" class="px-2.5 py-1 text-xs bg-blue-50 hover:bg-blue-100 text-blue-700 rounded-lg font-semibold transition-colors cursor-pointer">
                Pilih Router
              </button>
              <button onclick="Settings.toggleActive('${techId}', ${isActive})" class="px-2 py-1 text-xs border border-slate-200 hover:bg-slate-100 text-slate-700 rounded-lg font-medium transition-colors cursor-pointer">
                ${isActive ? 'Nonaktifkan' : 'Aktifkan'}
              </button>
            </div>
          </td>
        </tr>
      `;
    }).join('');
  },

  bindModals() {
    // Modal Tambah Teknisi
    const addBtn = document.getElementById('open-add-tech-btn');
    const addModal = document.getElementById('add-tech-modal');
    const closeAddBtn = document.getElementById('close-add-modal-btn');
    const addForm = document.getElementById('add-tech-form');

    if (addBtn && addModal) {
      addBtn.onclick = () => {
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
        const username = document.getElementById('tech-username').value.trim();
        const email = document.getElementById('tech-email').value.trim();
        const password = document.getElementById('tech-password').value;
        const fullName = document.getElementById('tech-fullname').value.trim();
        const phoneNumber = document.getElementById('tech-phone').value.trim();

        const selectedDevices = Array.from(
          document.querySelectorAll('#add-router-checkboxes input:checked')
        ).map(cb => cb.value);

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
          await this.loadTechnicians();
        } catch (err) {
          showToast('Gagal membuat teknisi: ' + err.message, 'error');
        } finally {
          if (submitBtn) submitBtn.disabled = false;
        }
      };
    }

    // Modal Assign Router
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
        if (!this.editingUser) return;

        const selectedDevices = Array.from(
          document.querySelectorAll('#edit-router-checkboxes input:checked')
        ).map(cb => cb.value);

        saveAssignBtn.disabled = true;

        try {
          await API.call('/polyglot.v1.UserService/AssignDevicesToUser', {
            userId: Number(this.editingUser.id),
            deviceIds: selectedDevices
          });

          showToast('Akses router teknisi berhasil diperbarui!', 'success');
          assignModal.classList.add('hidden');
          await this.loadTechnicians();
        } catch (err) {
          showToast('Gagal memperbarui router teknisi: ' + err.message, 'error');
        } finally {
          saveAssignBtn.disabled = false;
        }
      };
    }

    // Global ESC key listener to close open modals
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

    if (!this.adminDevices || this.adminDevices.length === 0) {
      container.innerHTML = '<p class="text-xs text-slate-400 italic p-3 text-center">Belum ada router di inventory sistem. Tambahkan router di Portal Utama terlebih dahulu.</p>';
      return;
    }

    const preselectedSet = new Set((preselectedIds || []).map(String));

    container.innerHTML = this.adminDevices.map(d => {
      const isChecked = preselectedSet.has(String(d.id));
      return `
        <label class="flex items-center gap-2.5 p-2.5 rounded-xl border border-slate-200 hover:bg-slate-100/80 cursor-pointer text-xs font-medium text-slate-800 transition-colors">
          <input type="checkbox" value="${d.id}" ${isChecked ? 'checked' : ''} class="w-4 h-4 rounded border-slate-300 text-blue-600 focus:ring-blue-500 cursor-pointer">
          <div class="flex flex-col">
            <span class="font-bold text-slate-900">${d.name || d.host}</span>
            <span class="text-[10px] text-slate-500 font-mono">${d.host}${d.port ? ':' + d.port : ''}</span>
          </div>
        </label>
      `;
    }).join('');
  },

  openAssignModal(techId) {
    const tech = this.technicians.find(t => String(t.id) === String(techId));
    if (!tech) {
      console.warn('Technician not found for id:', techId);
      return;
    }

    this.editingUser = tech;
    const modal = document.getElementById('assign-modal');
    const targetName = document.getElementById('assign-target-name');

    if (targetName) targetName.textContent = tech.fullName || tech.full_name || tech.username;
    this.renderRouterCheckboxes('edit-router-checkboxes', tech.assignedDeviceIds || tech.assigned_device_ids || []);

    if (modal) modal.classList.remove('hidden');
  },

  async toggleActive(techId, currentActive) {
    try {
      await API.call('/polyglot.v1.UserService/ToggleActive', {
        id: Number(techId)
      });
      showToast(`Status teknisi berhasil diubah`, 'success');
      await this.loadTechnicians();
    } catch (err) {
      showToast('Gagal mengubah status: ' + err.message, 'error');
    }
  }
};

window.Settings = Settings;
