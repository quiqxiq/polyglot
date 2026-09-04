/**
 * Polyglot Web Dial Router Selector
 * Mengelola router MikroTik aktif dan sinkronisasi ke seluruh komponen UI
 */

const RouterSelector = {
  devices: [],
  selectedDeviceId: localStorage.getItem('selected_device_id') || null,

  async init() {
    try {
      const res = await API.call('/polyglot.v1.DeviceService/ListDevices', {});
      this.devices = res.devices || [];

      if (this.devices.length === 0) {
        this.renderNoDevices();
        return;
      }

      // Periksa device yang tersimpan sebelumnya di localStorage
      const savedId = localStorage.getItem('selected_device_id');
      const found = this.devices.find(d => d.id === savedId);

      if (found) {
        this.selectedDeviceId = found.id;
      } else {
        this.selectedDeviceId = this.devices[0].id;
        localStorage.setItem('selected_device_id', this.selectedDeviceId);
      }

      this.renderDropdowns();
      this.notifyChange();
    } catch (err) {
      console.error('Failed to load devices:', err);
      showToast('Gagal memuat daftar router yang diizinkan', 'error');
    }
  },

  renderDropdowns() {
    const selects = document.querySelectorAll('.router-select-control');
    selects.forEach(select => {
      select.innerHTML = '';
      this.devices.forEach(dev => {
        const opt = document.createElement('option');
        opt.value = dev.id;
        opt.textContent = `${dev.name || dev.host} (${dev.host})`;
        if (dev.id === this.selectedDeviceId) {
          opt.selected = true;
        }
        select.appendChild(opt);
      });

      select.onchange = (e) => {
        this.setDevice(e.target.value);
      };
    });
  },

  renderNoDevices() {
    const selects = document.querySelectorAll('.router-select-control');
    selects.forEach(select => {
      select.innerHTML = '<option value="">Tidak ada router yang diizinkan</option>';
      select.disabled = true;
    });
    showToast('Akun Anda belum memiliki akses ke router MikroTik manapun. Hubungi Admin.', 'warning');
  },

  setDevice(deviceId) {
    if (this.selectedDeviceId === deviceId) return;
    this.selectedDeviceId = deviceId;
    localStorage.setItem('selected_device_id', deviceId);

    // Sync all dropdowns on page
    document.querySelectorAll('.router-select-control').forEach(select => {
      select.value = deviceId;
    });

    this.notifyChange();
  },

  notifyChange() {
    const currentDevice = this.getSelectedDevice();
    window.dispatchEvent(new CustomEvent('deviceChanged', {
      detail: {
        deviceId: this.selectedDeviceId,
        device: currentDevice
      }
    }));
  },

  getSelectedDeviceId() {
    return this.selectedDeviceId || localStorage.getItem('selected_device_id');
  },

  getSelectedDevice() {
    return this.devices.find(d => d.id === this.selectedDeviceId) || null;
  },

  onDeviceChange(callback) {
    window.addEventListener('deviceChanged', (e) => {
      callback(e.detail.deviceId, e.detail.device);
    });
    // Panggil langsung jika device sudah ada
    if (this.selectedDeviceId) {
      callback(this.selectedDeviceId, this.getSelectedDevice());
    }
  }
};

window.RouterSelector = RouterSelector;
