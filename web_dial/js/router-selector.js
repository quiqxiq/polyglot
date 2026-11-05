/**
 * Polyglot Web Dial Router Selector
 * Mengelola router MikroTik aktif dan sinkronisasi ke seluruh komponen UI
 */

const RouterSelector = {
  devices: [],
  selectedDeviceId: localStorage.getItem('selected_device_id') || null,

  async init() {
    // Router inventory is readable with dial:device:read or device:read.
    if (typeof Auth !== 'undefined' && !Auth.can('dial:device:read') && !Auth.can('device:read')) {
      this.devices = [];
      return false;
    }

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
      return true;
    } catch (err) {
      console.error('Failed to load devices:', err);
      this.devices = [];
      this.renderNoDevices();
      const errMsg = typeof I18n !== 'undefined' ? I18n.t('failed_load_devices') : 'Gagal memuat daftar router yang diizinkan';
      API.reportError(err, errMsg);
      return false;
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
    const noDevMsg = typeof I18n !== 'undefined'
      ? (I18n.getLanguage() === 'en' ? 'No authorized routers' : 'Tidak ada router yang diizinkan')
      : 'Tidak ada router yang diizinkan';
    const selects = document.querySelectorAll('.router-select-control');
    selects.forEach(select => {
      select.innerHTML = `<option value="">${noDevMsg}</option>`;
      select.disabled = true;
    });
    const warningMsg = typeof I18n !== 'undefined'
      ? (I18n.getLanguage() === 'en' ? 'Your account does not have access to any MikroTik router. Contact Admin.' : 'Akun Anda belum memiliki akses ke router MikroTik manapun. Hubungi Admin.')
      : 'Akun Anda belum memiliki akses ke router MikroTik manapun. Hubungi Admin.';
    showToast(warningMsg, 'warning');
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
