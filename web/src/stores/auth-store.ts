import { create } from 'zustand'

interface AuthUser {
  accountNo: string
  email: string
  role: string[]
  // Permissions efektif dari backend (GetMe/Login) — diflatten ke format
  // "resource:action" dengan wildcard regex, mis. "knowledge:.*:*".
  permissions: string[]
  exp: number
}

interface AuthState {
  auth: {
    user: AuthUser | null
    setUser: (user: AuthUser | null) => void
    accessToken: string
    setAccessToken: (accessToken: string) => void
    resetAccessToken: () => void
    reset: () => void
  }
}

// Access token disimpan DI MEMORY saja (bukan cookie) — short-lived (default
// 1 jam) dan hilang saat reload halaman. Sesi bertahan lewat refresh token
// httpOnly (cookie `polyglot_refresh` yang di-set server): request membawa
// cookie otomatis (credentials: include), dan interceptor/api-client akan
// silent-refresh ketika token tidak ada / expired. Menyimpan token di cookie
// yang bisa dibaca JS adalah celah XSS token-theft — itulah yang dihindari.
export const useAuthStore = create<AuthState>()((set) => ({
  auth: {
    user: null,
    setUser: (user) =>
      set((state) => ({ ...state, auth: { ...state.auth, user } })),
    accessToken: '',
    setAccessToken: (accessToken) =>
      set((state) => ({ ...state, auth: { ...state.auth, accessToken } })),
    resetAccessToken: () =>
      set((state) => ({ ...state, auth: { ...state.auth, accessToken: '' } })),
    reset: () =>
      set((state) => ({
        ...state,
        auth: { ...state.auth, user: null, accessToken: '' },
      })),
  },
}))
