/**
 * Polyglot ConnectRPC API Client
 * Mengirimkan request JSON over HTTP ke endpoint ConnectRPC Polyglot
 */

const API = {
  // Base URL: Kosong jika reverse proxy (same-origin /polyglot.v1...),
  // atau ambil dari window.POLYGLOT_API_URL jika di-override.
  baseUrl: window.POLYGLOT_API_URL || '',

  /**
   * Mengirimkan ConnectRPC POST request
   * @param {string} servicePath e.g. '/polyglot.v1.PPPService/ListActiveSessions'
   * @param {object} payload Request message body
   * @param {boolean} requiresAuth Apakah membutuhkan token Bearer (default true)
   */
  async call(servicePath, payload = {}, requiresAuth = true) {
    const url = `${this.baseUrl}${servicePath}`;
    const headers = {
      'Content-Type': 'application/json',
    };

    if (requiresAuth) {
      const token = localStorage.getItem('polyglot_token');
      if (!token) {
        this.redirectToLogin('Silakan login terlebih dahulu');
        throw new Error('Unauthenticated');
      }
      headers['Authorization'] = `Bearer ${token}`;
    }

    try {
      const response = await fetch(url, {
        method: 'POST',
        headers,
        body: JSON.stringify(payload)
      });

      if (response.status === 401 && requiresAuth) {
        localStorage.removeItem('polyglot_token');
        localStorage.removeItem('polyglot_user');
        this.redirectToLogin('Sesi Anda telah berakhir, silakan login kembali');
        throw new Error('Sesi telah berakhir');
      }

      const data = await response.json().catch(() => ({}));

      if (!response.ok) {
        const errorMsg = data.message || `Request failed with status ${response.status}`;
        const err = new Error(errorMsg);
        err.code = data.code;
        err.status = response.status;
        throw err;
      }

      return data;
    } catch (err) {
      console.error(`API Error on ${servicePath}:`, err);
      throw err;
    }
  },

  /**
   * Mengirimkan ConnectRPC streaming request dan mengurai Connect framing envelope
   * Format envelope (5 bytes header):
   * - Byte 0: Flags (0 = data frame, 2 = end-of-stream trailer frame)
   * - Bytes 1-4: Length (32-bit unsigned big-endian integer)
   * - Next <Length> bytes: JSON string payload
   *
   * @param {string} servicePath e.g. '/polyglot.v1.DeviceService/StreamInterfaceTraffic'
   * @param {object} payload Request message body
   * @param {function} onMessage Callback untuk setiap frame data
   * @param {function} onError Callback jika terjadi error
   * @param {AbortSignal} [signal] Optional AbortSignal untuk membatalkan stream
   * @param {boolean} [requiresAuth=true]
   */
  async stream(servicePath, payload = {}, onMessage = () => {}, onError = () => {}, signal = null, requiresAuth = true) {
    const url = `${this.baseUrl}${servicePath}`;
    const headers = {
      'Content-Type': 'application/connect+json',
      'Connect-Protocol-Version': '1',
    };

    if (requiresAuth) {
      const token = localStorage.getItem('polyglot_token');
      if (!token) {
        this.redirectToLogin('Silakan login terlebih dahulu');
        const err = new Error('Unauthenticated');
        onError(err);
        throw err;
      }
      headers['Authorization'] = `Bearer ${token}`;
    }

    // Connect protocol streaming mewajibkan request body dikemas dalam framing envelope:
    // 1 byte flags (0) + 4 bytes big-endian length + JSON payload UTF-8
    const payloadBytes = new TextEncoder().encode(JSON.stringify(payload));
    const envelopedBody = new Uint8Array(5 + payloadBytes.length);
    envelopedBody[0] = 0;
    envelopedBody[1] = (payloadBytes.length >> 24) & 0xff;
    envelopedBody[2] = (payloadBytes.length >> 16) & 0xff;
    envelopedBody[3] = (payloadBytes.length >> 8) & 0xff;
    envelopedBody[4] = payloadBytes.length & 0xff;
    envelopedBody.set(payloadBytes, 5);

    try {
      const response = await fetch(url, {
        method: 'POST',
        headers,
        body: envelopedBody,
        signal
      });

      if (response.status === 401 && requiresAuth) {
        localStorage.removeItem('polyglot_token');
        localStorage.removeItem('polyglot_user');
        this.redirectToLogin('Sesi Anda telah berakhir, silakan login kembali');
        const err = new Error('Sesi telah berakhir');
        if (typeof onError === 'function') onError(err);
        return;
      }

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        const errorMsg = errorData.message || `Stream failed with status ${response.status}`;
        const err = new Error(errorMsg);
        err.status = response.status;
        err.code = errorData.code;
        if (typeof onError === 'function') onError(err);
        return;
      }

      if (!response.body) {
        const err = new Error('ReadableStream tidak didukung pada browser ini');
        if (typeof onError === 'function') onError(err);
        return;
      }

      const reader = response.body.getReader();
      let buffer = new Uint8Array(0);

      while (true) {
        let readResult;
        try {
          readResult = await reader.read();
        } catch (readErr) {
          // Firefox throws TypeError: Error in input stream when stream closes/navigates
          if (
            readErr.name === 'AbortError' ||
            signal?.aborted ||
            (readErr.message && (
              readErr.message.includes('input stream') ||
              readErr.message.includes('aborted') ||
              readErr.message.includes('network error')
            ))
          ) {
            return;
          }
          throw readErr;
        }

        const { done, value } = readResult;
        if (done) break;

        const newBuf = new Uint8Array(buffer.length + value.length);
        newBuf.set(buffer, 0);
        newBuf.set(value, buffer.length);
        buffer = newBuf;

        while (buffer.length >= 5) {
          const flags = buffer[0];
          const length = (buffer[1] << 24) | (buffer[2] << 16) | (buffer[3] << 8) | buffer[4];

          if (buffer.length < 5 + length) {
            break;
          }

          const payloadBytes = buffer.slice(5, 5 + length);
          buffer = buffer.slice(5 + length);

          if (flags & 2) {
            try {
              const trailerText = new TextDecoder().decode(payloadBytes);
              const trailer = JSON.parse(trailerText);
              if (trailer.error && typeof onError === 'function') {
                const err = new Error(trailer.error.message || 'Stream error from server');
                err.code = trailer.error.code;
                onError(err);
              }
            } catch (_) {}
            continue;
          }

          try {
            const jsonText = new TextDecoder().decode(payloadBytes);
            if (jsonText.trim() && typeof onMessage === 'function') {
              const msg = JSON.parse(jsonText);
              onMessage(msg);
            }
          } catch (jsonErr) {
            console.warn('Failed to parse stream frame JSON:', jsonErr);
          }
        }
      }
    } catch (err) {
      if (
        err.name === 'AbortError' ||
        signal?.aborted ||
        (err.message && (
          err.message.includes('input stream') ||
          err.message.includes('aborted') ||
          err.message.includes('network error')
        ))
      ) {
        return;
      }
      console.warn(`Stream warning on ${servicePath}:`, err.message || err);
      if (typeof onError === 'function') {
        onError(err);
      }
    }
  },

  /**
   * Helper untuk membaca 1 frame data pertama dari stream lalu langsung membatalkan koneksi
   * Sangat berguna untuk mengambil snapshot on-demand (seperti daftar Simple Queues)
   */
  async streamOnce(servicePath, payload = {}, requiresAuth = true) {
    const controller = new AbortController();
    return new Promise((resolve, reject) => {
      let resolved = false;
      this.stream(
        servicePath,
        payload,
        (msg) => {
          if (!resolved) {
            resolved = true;
            controller.abort();
            resolve(msg);
          }
        },
        (err) => {
          if (!resolved) {
            resolved = true;
            reject(err);
          }
        },
        controller.signal,
        requiresAuth
      ).catch((err) => {
        if (!resolved) {
          resolved = true;
          reject(err);
        }
      });
    });
  },

  redirectToLogin(msg) {
    if (window.location.pathname.endsWith('login.html')) return;
    if (msg) sessionStorage.setItem('login_notice', msg);
    window.location.href = 'login.html';
  }
};

// Global UI Utilities (Aligned with gnet_dial design)
const UIUtils = {
  getProfileInitials(name) {
    if (!name || name === 'N/A' || name === '-') return 'NA';
    const words = name.trim().split(/[\s._-]+/).filter(Boolean);
    if (words.length >= 2) {
      return (words[0][0] + words[1][0]).toUpperCase();
    }
    const clean = name.trim();
    if (clean.length >= 2) {
      return (clean[0] + clean[clean.length - 1]).toUpperCase();
    }
    return (clean[0] || 'U').toUpperCase();
  },

  getAvatarColor(name) {
    const colors = [
      'bg-blue-600', 'bg-emerald-600', 'bg-purple-600', 'bg-amber-600', 'bg-rose-600',
      'bg-indigo-600', 'bg-cyan-600', 'bg-teal-600', 'bg-orange-600', 'bg-violet-600'
    ];
    let hash = 0;
    const str = name || 'default';
    for (let i = 0; i < str.length; i++) {
      hash = ((hash << 5) - hash) + str.charCodeAt(i);
      hash = hash & hash;
    }
    return colors[Math.abs(hash) % colors.length];
  },

  formatUptime(uptime) {
    if (!uptime || uptime === 'N/A' || uptime === '-' || uptime === '0s') return '-';
    let days = 0, hours = 0, minutes = 0, seconds = 0;

    // Check RouterOS formats: e.g. "2d01:17:03", "01:17:03", "2d1h17m3s", "15m20s"
    const colonMatch = uptime.match(/(?:(\d+)d)?(?:(\d+)h)?(\d{2}):(\d{2}):(\d{2})/);
    if (colonMatch) {
      if (colonMatch[1]) days = parseInt(colonMatch[1], 10);
      if (colonMatch[2]) hours = parseInt(colonMatch[2], 10);
      hours = parseInt(colonMatch[3], 10);
      minutes = parseInt(colonMatch[4], 10);
      seconds = parseInt(colonMatch[5], 10);
    } else {
      const dMatch = uptime.match(/(\d+)d/);
      const hMatch = uptime.match(/(\d+)h/);
      const mMatch = uptime.match(/(\d+)m/);
      const sMatch = uptime.match(/(\d+)s/);
      if (dMatch) days = parseInt(dMatch[1], 10);
      if (hMatch) hours = parseInt(hMatch[1], 10);
      if (mMatch) minutes = parseInt(mMatch[1], 10);
      if (sMatch) seconds = parseInt(sMatch[1], 10);
    }

    const parts = [];
    if (days > 0) parts.push(`${days} hari`);
    const hStr = hours.toString().padStart(2, '0');
    const mStr = minutes.toString().padStart(2, '0');
    const sStr = seconds.toString().padStart(2, '0');
    parts.push(`${hStr}:${mStr}:${sStr}`);
    return parts.join(' ');
  },

  formatLastLogout(datetime) {
    if (!datetime || datetime === 'N/A' || datetime === '-' || datetime === '') return '-';
    try {
      // RouterOS format e.g. "mar/04/2026 18:23:45" or "2026-03-04 18:23:45"
      let d = new Date(datetime.replace(' ', 'T'));
      if (isNaN(d.getTime())) {
        // Parse "mmm/dd/yyyy hh:mm:ss"
        const parts = datetime.split(/[\s/:]+/);
        if (parts.length >= 6) {
          const monthNames = { jan: 0, feb: 1, mar: 2, apr: 3, may: 4, jun: 5, jul: 6, aug: 7, sep: 8, oct: 9, nov: 10, dec: 11 };
          const m = monthNames[parts[0].toLowerCase()] ?? 0;
          d = new Date(parseInt(parts[2]), m, parseInt(parts[1]), parseInt(parts[3]), parseInt(parts[4]), parseInt(parts[5] || 0));
        }
      }
      if (isNaN(d.getTime())) return datetime;

      const months = ['Jan', 'Feb', 'Mar', 'Apr', 'Mei', 'Jun', 'Jul', 'Agu', 'Sep', 'Okt', 'Nov', 'Des'];
      const day = d.getDate();
      const month = months[d.getMonth()];
      const year = d.getFullYear();
      const hours = d.getHours().toString().padStart(2, '0');
      const minutes = d.getMinutes().toString().padStart(2, '0');
      return `${day} ${month} ${year}, ${hours}:${minutes}`;
    } catch {
      return datetime;
    }
  },

  getProfileIcon(profile = '', extraClass = '') {
    const p = (profile || '').toLowerCase();
    const cls = `lucide ${extraClass}`;
    if (p.includes('pppoe')) {
      return `<svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="${cls}"><rect x="16" y="16" width="6" height="6" rx="1"/><rect x="2" y="16" width="6" height="6" rx="1"/><rect x="9" y="2" width="6" height="6" rx="1"/><path d="M5 16v-3a1 1 0 0 1 1-1h12a1 1 0 0 1 1 1v3"/><path d="M12 12V8"/></svg>`;
    } else if (p.includes('hotspot')) {
      return `<svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="${cls}"><path d="M5 12.55a11 11 0 0 1 14.08 0"/><path d="M1.42 9a16 16 0 0 1 21.16 0"/><path d="M8.53 16.11a6 6 0 0 1 6.95 0"/><line x1="12" y1="20" x2="12.01" y2="20"/></svg>`;
    } else if (p.includes('vpn')) {
      return `<svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="${cls}"><rect width="18" height="11" x="3" y="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>`;
    }
    return `<svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="${cls}"><path d="M19 21v-2a4 4 0 0 0-4-4H9a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>`;
  }
};

// Global click listener to toggle and close action menus (three-dot mobile menu)
document.addEventListener('click', (e) => {
  const btn = e.target.closest('.action-menu-btn');
  if (btn) {
    e.stopPropagation();
    const menu = btn.parentElement.querySelector('.action-menu');
    document.querySelectorAll('.action-menu').forEach(m => {
      if (m !== menu) m.classList.add('hidden');
    });
    if (menu) menu.classList.toggle('hidden');
    return;
  }
  // Click outside closes all open menus
  document.querySelectorAll('.action-menu').forEach(m => m.classList.add('hidden'));
});

// Expose to window
window.API = API;
window.UIUtils = UIUtils;

