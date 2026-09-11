import { useMemo, useState } from 'react'
import { CreditCard, Loader2, Save } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import {
  BatchUpdateSettingsRequest,
  type SettingItem,
} from '@/gen/v1/settings_pb'
import {
  useBatchUpdateSettingsMutation,
  useSettingsByCategoryQuery,
} from '../api/use-settings'

const SECRET_SUFFIXES = [
  'api_key',
  'private_key',
  'secret_key',
  'server_key',
  'callback_token',
]

const KNOWN_GROUPS = ['active', 'gw.tripay', 'gw.midtrans', 'gw.xendit']

function isSecretKey(key: string) {
  return SECRET_SUFFIXES.some(
    (suffix) => key === suffix || key.endsWith(`.${suffix}`) || key.endsWith(`_${suffix}`)
  )
}

function groupLabel(group: string) {
  switch (group) {
    case 'active':
      return 'Gateway Aktif'
    case 'gw.tripay':
      return 'Tripay'
    case 'gw.midtrans':
      return 'Midtrans'
    case 'gw.xendit':
      return 'Xendit'
    default:
      return group
  }
}

function groupDescription(group: string) {
  if (group === 'active') {
    return 'Pilih gateway default untuk tagihan portal mandiri maupun kasir.'
  }
  return 'Aktifkan dan isi kredensial gateway. Kredensial sensitif disimpan terenkripsi (AES-GCM) oleh backend.'
}

export function GatewaySettingsForm() {
  const { data, isLoading } = useSettingsByCategoryQuery('isp_gateway')
  const batch = useBatchUpdateSettingsMutation()
  // Hanya menyimpan hasil edit; nilai dasar dibaca langsung dari server.
  const [edits, setEdits] = useState<Record<string, string>>({})

  const groups = useMemo(() => {
    const map = new Map<string, SettingItem[]>()
    for (const item of data ?? []) {
      const group =
        item.key === 'gw.active' ? 'active' : item.key.split('.').slice(0, 2).join('.')
      const list = map.get(group) ?? []
      list.push(item)
      map.set(group, list)
    }
    return [...map.entries()]
      .filter(([group]) => KNOWN_GROUPS.includes(group))
      .sort(([a], [b]) => a.localeCompare(b))
  }, [data])

  const valueOf = (item: SettingItem) => edits[item.key] ?? item.value

  const setValue = (key: string, value: string) => {
    setEdits((prev) => ({ ...prev, [key]: value }))
  }

  const onSave = () => {
    const settings = (data ?? []).map((item) => ({
      key: item.key,
      value: edits[item.key] ?? item.value,
      category: 'isp_gateway',
      description: item.description,
    }))
    batch.mutate(new BatchUpdateSettingsRequest({ settings }))
  }

  const renderField = (item: SettingItem) => {
    const value = valueOf(item)
    const label = item.description || item.key

    if (item.key === 'gw.active') {
      return (
        <div key={item.key} className='space-y-2'>
          <Label>{label}</Label>
          <Select value={value || 'TRIPAY'} onValueChange={(v) => setValue(item.key, v)}>
            <SelectTrigger className='w-full md:w-72'>
              <SelectValue placeholder='Pilih gateway' />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value='TRIPAY'>Tripay</SelectItem>
              <SelectItem value='MIDTRANS'>Midtrans</SelectItem>
              <SelectItem value='XENDIT'>Xendit</SelectItem>
            </SelectContent>
          </Select>
          <p className='text-xs text-muted-foreground'>{item.key}</p>
        </div>
      )
    }

    if (item.key.endsWith('.enabled')) {
      return (
        <div
          key={item.key}
          className='flex flex-row items-center justify-between rounded-lg border p-3'
        >
          <div className='space-y-0.5'>
            <Label className='text-sm font-semibold'>{label}</Label>
            <p className='text-xs text-muted-foreground'>{item.key}</p>
          </div>
          <Switch
            checked={value === 'true'}
            onCheckedChange={(checked) => setValue(item.key, checked ? 'true' : 'false')}
          />
        </div>
      )
    }

    return (
      <div key={item.key} className='space-y-2'>
        <Label>{label}</Label>
        <Input
          type={isSecretKey(item.key) ? 'password' : 'text'}
          value={value}
          placeholder={item.key}
          autoComplete='off'
          onChange={(e) => setValue(item.key, e.target.value)}
        />
        <p className='text-xs text-muted-foreground'>{item.key}</p>
      </div>
    )
  }

  if (isLoading) {
    return (
      <div className='flex h-48 items-center justify-center'>
        <Loader2 className='h-8 w-8 animate-spin text-muted-foreground' />
      </div>
    )
  }

  return (
    <div className='space-y-6'>
      {groups.map(([group, items]) => (
        <Card key={group}>
          <CardHeader className='pb-3'>
            <CardTitle className='flex items-center gap-2 text-base'>
              <CreditCard className='h-5 w-5 text-primary' />
              {groupLabel(group)}
            </CardTitle>
            <CardDescription>{groupDescription(group)}</CardDescription>
          </CardHeader>
          <CardContent className='space-y-4'>{items.map(renderField)}</CardContent>
        </Card>
      ))}

      <Button
        type='button'
        disabled={batch.isPending}
        className='min-w-44'
        onClick={onSave}
      >
        {batch.isPending ? (
          <Loader2 className='mr-2 h-4 w-4 animate-spin' />
        ) : (
          <Save className='mr-2 h-4 w-4' />
        )}
        Simpan Konfigurasi Gateway
      </Button>
    </div>
  )
}
