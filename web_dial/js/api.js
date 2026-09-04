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

// Expose to window
window.API = API;

