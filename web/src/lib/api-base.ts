// Base URL API.
// - Nilai eksplisit (mis. http://localhost:8080 atau https://domain): dipakai apa adanya.
// - "/" atau kosong: same-origin — di produksi request lewat gateway nginx
//   yang mem-proxy /polyglot.v1. dan /api/ ke backend.
export const API_BASE_URL = (() => {
  const raw = import.meta.env.VITE_API_URL
  if (raw && raw !== '/') {
    return raw.replace(/\/+$/, '')
  }
  return import.meta.env.DEV ? 'http://localhost:8080' : ''
})()
