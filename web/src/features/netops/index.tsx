import { useEffect, useState } from 'react'
import { Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useI18n } from '@/context/i18n-provider'
import { authClient } from '@/lib/api-client'
import { API_BASE_URL } from '@/lib/api-base'
import { useAuthStore } from '@/stores/auth-store'

interface TicketResponse {
  url?: string
}

async function requestTicket(): Promise<string> {
  let token = useAuthStore.getState().auth.accessToken
  if (!token) {
    const refreshed = await authClient.refreshToken({})
    if (!refreshed.token) throw new Error('missing access token')
    token = refreshed.token
    useAuthStore.getState().auth.setAccessToken(token)
  }

  const res = await fetch(`${API_BASE_URL}/api/netops/ticket`, {
    method: 'POST',
    credentials: 'include',
    headers: { Authorization: `Bearer ${token}` },
  })
  if (!res.ok) throw new Error(`ticket request failed with ${res.status}`)

  const body = (await res.json()) as TicketResponse
  if (!body.url) throw new Error('ticket response has no url')
  return body.url
}

// NetOpsLaunch mints a one-time ticket and hands the browser to the Go
// reverse proxy. The Polyglot JWT never appears in the NetOps URL.
export function NetOpsLaunch() {
  const { t } = useI18n()
  const [failed, setFailed] = useState(false)
  const [attempt, setAttempt] = useState(0)

  useEffect(() => {
    let cancelled = false
    const launch = async () => {
      try {
        const target = await requestTicket()
        if (cancelled) return
        const base = API_BASE_URL || window.location.origin
        window.location.assign(new URL(target, base).href)
      } catch {
        if (!cancelled) setFailed(true)
      }
    }
    void launch()
    return () => {
      cancelled = true
    }
  }, [attempt])

  return (
    <div className='flex h-svh flex-col items-center justify-center gap-4 p-6 text-center'>
      {failed ? (
        <>
          <p className='max-w-md text-sm text-muted-foreground'>
            {t('netops.failed')}
          </p>
          <Button
            type='button'
            onClick={() => {
              setFailed(false)
              setAttempt((value) => value + 1)
            }}
          >
            {t('netops.retry')}
          </Button>
        </>
      ) : (
        <>
          <Loader2 className='size-6 animate-spin text-muted-foreground' />
          <div className='space-y-1'>
            <h1 className='text-lg font-semibold'>{t('netops.opening')}</h1>
            <p className='max-w-md text-sm text-muted-foreground'>
              {t('netops.opening_desc')}
            </p>
          </div>
        </>
      )}
    </div>
  )
}
