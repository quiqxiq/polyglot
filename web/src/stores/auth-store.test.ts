import { beforeEach, describe, expect, it } from 'vitest'
import { useAuthStore } from './auth-store'

const sampleUser = {
  accountNo: 'ACC-1',
  email: 'user@example.com',
  role: ['user'],
  exp: 1_700_000_000,
}

describe('useAuthStore', () => {
  beforeEach(() => {
    // Fresh state per test; the store itself holds no persisted storage.
    useAuthStore.getState().auth.reset()
    localStorage.clear()
    document.cookie.split(';').forEach((c) => {
      document.cookie = c
        .replace(/^ +/, '')
        .replace(/=.*/, '=;expires=' + new Date(0).toUTCString() + ';path=/')
    })
  })

  it('starts with an empty access token', () => {
    expect(useAuthStore.getState().auth.accessToken).toBe('')
    expect(useAuthStore.getState().auth.user).toBeNull()
  })

  it('keeps the access token in memory only — never persisted to storage', () => {
    useAuthStore.getState().auth.setAccessToken('session-token')

    expect(useAuthStore.getState().auth.accessToken).toBe('session-token')
    // Memory-only means the token must not leak into any persisted storage,
    // so a hard reload (localStorage/cookie) cannot restore it.
    expect(localStorage.getItem('accessToken')).toBeNull()
    expect(localStorage.getItem('auth')).toBeNull()
    expect(document.cookie).not.toContain('session-token')
  })

  it('clears access token when resetAccessToken is used', () => {
    useAuthStore.getState().auth.setAccessToken('to-clear')
    useAuthStore.getState().auth.resetAccessToken()

    expect(useAuthStore.getState().auth.accessToken).toBe('')
  })

  it('updates the signed-in user via setUser', () => {
    useAuthStore.getState().auth.setUser({ ...sampleUser })

    expect(useAuthStore.getState().auth.user).toEqual(sampleUser)
  })

  it('reset clears user and access token', () => {
    useAuthStore.getState().auth.setAccessToken('will-be-cleared')
    useAuthStore.getState().auth.setUser({ ...sampleUser })

    useAuthStore.getState().auth.reset()

    expect(useAuthStore.getState().auth.user).toBeNull()
    expect(useAuthStore.getState().auth.accessToken).toBe('')
  })
})
