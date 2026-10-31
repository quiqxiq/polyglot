import { useState } from 'react'
import { Link } from '@tanstack/react-router'
import { AlertCircleIcon, BotIcon, ShieldCheckIcon } from 'lucide-react'
import { useDeviceStore } from '@/stores/device-store'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { ConfigDrawer } from '@/components/config-drawer'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { useDevicesQuery } from '@/features/devices/api/use-devices'
import { useLLMConfigsQuery } from '@/features/skills/api/use-llm-config'
import { useAiOpsChat } from '../api/use-ai-ops-chat'
import { ChatComposer } from './chat-composer'
import { ChatThread } from './chat-thread'

export function AIOpsChatPage() {
  const { selectedDeviceId } = useDeviceStore()
  const { data: devices = [] } = useDevicesQuery()
  const { data: configs = [] } = useLLMConfigsQuery()
  const [modelConfigId, setModelConfigId] = useState('')
  const [input, setInput] = useState('')

  const device = devices.find((item) => item.id === selectedDeviceId)
  const chat = useAiOpsChat({ deviceId: selectedDeviceId, modelConfigId })
  const isStreaming = chat.status === 'streaming'

  function send() {
    const text = input.trim()
    if (!text || chat.isBusy || !selectedDeviceId) return
    setInput('')
    void chat.send(text)
  }

  return (
    <>
      <Header fixed>
        <Search className='me-auto' />
        <ThemeSwitch />
        <ConfigDrawer />
        <ProfileDropdown />
      </Header>
      <Main className='flex flex-1 flex-col gap-4 sm:gap-6'>
        <div className='flex flex-wrap items-end justify-between gap-2'>
          <div>
            <div className='flex items-center gap-2'>
              <BotIcon className='h-6 w-6 text-primary' />
              <h1 className='text-2xl font-bold tracking-tight'>
                AI Router Ops
              </h1>
            </div>
            <p className='mt-0.5 text-sm text-muted-foreground'>
              {device
                ? `Target: ${device.name} (${device.host})`
                : 'Pilih router di sidebar untuk mulai.'}{' '}
              · READ bebas, MUTATION wajib approval.
            </p>
          </div>
          <div className='flex items-center gap-2'>
            <Select
              value={modelConfigId || 'ACTIVE'}
              onValueChange={(value) =>
                setModelConfigId(value === 'ACTIVE' ? '' : value)
              }
            >
              <SelectTrigger className='h-9 w-52'>
                <SelectValue placeholder='Model aktif' />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value='ACTIVE'>Model aktif</SelectItem>
                {configs.map((config) => (
                  <SelectItem key={config.id} value={config.id}>
                    {config.provider} · {config.modelName}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Button asChild variant='outline' size='sm' className='gap-1.5'>
              <Link to='/ai-ops/approvals'>
                <ShieldCheckIcon className='h-4 w-4' /> Approvals
              </Link>
            </Button>
          </div>
        </div>

        <Card className='flex min-h-[460px] flex-1 flex-col'>
          <CardContent className='flex flex-1 flex-col gap-3 pt-4'>
            <ChatThread
              messages={chat.messages}
              trace={chat.trace}
              isStreaming={isStreaming}
              onRetry={() => void chat.retry()}
              onSuggestion={setInput}
            />

            {chat.error && (
              <div
                role='alert'
                className='flex items-start gap-2 rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-xs text-destructive'
              >
                <AlertCircleIcon className='mt-0.5 size-3.5 shrink-0' />
                <div className='min-w-0 flex-1'>
                  <p className='font-medium'>Percakapan gagal</p>
                  <p className='break-words'>{chat.error.message}</p>
                </div>
                <Button
                  size='xs'
                  variant='outline'
                  onClick={() => void chat.retry()}
                >
                  Coba lagi
                </Button>
              </div>
            )}

            <ChatComposer
              value={input}
              onValueChange={setInput}
              onSend={send}
              onStop={chat.stop}
              busy={chat.isBusy}
              disabled={!selectedDeviceId}
            />
          </CardContent>
        </Card>

        <p className='text-[11px] text-muted-foreground'>
          AI tidak dapat menyetujui request sendiri. Approval selalu lewat sesi
          user dengan permission approval:decide.
        </p>
      </Main>
    </>
  )
}
