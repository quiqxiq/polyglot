import { useMemo, useState } from 'react'
import { Loader2, ReceiptText, Save } from 'lucide-react'
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
import { BatchUpdateSettingsRequest, SettingItem } from '@/gen/v1/settings_pb'
import {
  useBatchUpdateSettingsMutation,
  useSettingsByCategoryQuery,
} from '../api/use-settings'

interface FieldDef {
  key: string
  label: string
  description?: string
  type: 'text' | 'number' | 'switch' | 'select'
  options?: { value: string; label: string }[]
  min?: number
  max?: number
}

const BILLING_FIELDS: FieldDef[] = [
  {
    key: 'isp.billing_mode',
    label: 'Mode Billing',
    type: 'select',
    options: [
      { value: 'PASCABAYAR', label: 'Pascabayar (pakai dulu, bayar kemudian)' },
      { value: 'PRABAYAR', label: 'Prabayar (bayar dulu, baru pakai)' },
    ],
    description:
      'Prabayar menerbitkan invoice beberapa hari sebelum siklus dimulai sesuai lead days.',
  },
  {
    key: 'isp.closing_date',
    label: 'Tanggal Tutup Buku / Anchor',
    type: 'number',
    min: 1,
    max: 28,
    description: 'Menjadi default anchor langganan baru (1–28).',
  },
  {
    key: 'isp.prepay_invoice_lead_days',
    label: 'Lead Days Prabayar',
    type: 'number',
    min: 0,
    max: 30,
    description: 'Berapa hari sebelum siklus invoice prabayar diterbitkan.',
  },
  {
    key: 'isp.unpaid_days_display',
    label: 'Ambang Label "Belum Bayar" (hari)',
    type: 'number',
    min: 1,
    max: 90,
    description: 'Hanya tampilan/filter, bukan denda atau biaya.',
  },
  {
    key: 'isp.tax_percent',
    label: 'PPN Default Mitra (%)',
    type: 'number',
    min: 0,
    max: 100,
    description: 'Fallback bila paket tidak mengisi tarif pajak. 0 = tanpa PPN.',
  },
  {
    key: 'isp.reminder_pre_offsets',
    label: 'Offset Pengingat Sebelum Jatuh Tempo',
    type: 'text',
    description: 'Contoh: 7,3,1,0 (hari).',
  },
  {
    key: 'isp.reminder_post_offsets',
    label: 'Offset Pengingat Setelah Jatuh Tempo',
    type: 'text',
    description: 'Contoh: 1,3,7 (hari). Pengingat, bukan denda.',
  },
  {
    key: 'isp.debt_reminder_enabled',
    label: 'Pengingat Hutang WhatsApp',
    type: 'switch',
    description: 'Kirim pengingat untuk hutang yang mendekati jatuh tempo.',
  },
  {
    key: 'isp.debt_reminder_days',
    label: 'Offset Pengingat Hutang',
    type: 'text',
    description: 'Contoh: 3,1,0 (hari sebelum jatuh tempo).',
  },
  {
    key: 'isp.mass_payment_enabled',
    label: 'Aktifkan Pembayaran Massal',
    type: 'switch',
    description: 'Batch pelunasan banyak invoice sekaligus lewat antrean job.',
  },
]

const GATEWAY_OPS_FIELDS: FieldDef[] = [
  {
    key: 'isp.gateway_sync_stale_minutes',
    label: 'Sinkronisasi Transaksi PENDING (menit)',
    type: 'number',
    min: 1,
    max: 1440,
    description: 'Transaksi PENDING lebih tua dari ini akan disinkronkan/expired.',
  },
  {
    key: 'isp.gateway_fee_expense_category_id',
    label: 'Kategori Kas Beban Biaya Gateway',
    type: 'text',
    description: 'Default: cc-gateway-fee.',
  },
]

function toValueMap(items: SettingItem[] | undefined) {
  const map = new Map<string, string>()
  for (const item of items ?? []) map.set(item.key, item.value)
  return map
}

function FieldInput({
  def,
  value,
  onChange,
}: {
  def: FieldDef
  value: string
  onChange: (value: string) => void
}) {
  if (def.type === 'switch') {
    return (
      <div className='flex items-center gap-2'>
        <Switch
          checked={value === 'true'}
          onCheckedChange={(checked) => onChange(checked ? 'true' : 'false')}
        />
        <span className='text-xs text-muted-foreground'>
          {value === 'true' ? 'Aktif' : 'Nonaktif'}
        </span>
      </div>
    )
  }
  if (def.type === 'select') {
    return (
      <Select value={value} onValueChange={onChange}>
        <SelectTrigger>
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          {(def.options ?? []).map((opt) => (
            <SelectItem key={opt.value} value={opt.value}>
              {opt.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    )
  }
  return (
    <Input
      type={def.type === 'number' ? 'number' : 'text'}
      min={def.min}
      max={def.max}
      value={value}
      onChange={(e) => onChange(e.target.value)}
    />
  )
}

function CategoryCard({
  title,
  description,
  category,
  fields,
}: {
  title: string
  description: string
  category: string
  fields: FieldDef[]
}) {
  const { data, isLoading } = useSettingsByCategoryQuery(category)
  const batch = useBatchUpdateSettingsMutation()
  const initial = useMemo(() => toValueMap(data), [data])
  // Hanya menyimpan hasil edit; nilai dasar dibaca langsung dari server.
  const [overrides, setOverrides] = useState<Record<string, string>>({})

  const valueOf = (key: string) => overrides[key] ?? initial.get(key) ?? ''
  const dirty = fields.some((f) => valueOf(f.key) !== (initial.get(f.key) ?? ''))

  async function handleSave() {
    const items: SettingItem[] = []
    for (const f of fields) {
      const next = valueOf(f.key)
      if (next !== (initial.get(f.key) ?? '')) {
        items.push(
          new SettingItem({
            key: f.key,
            value: next,
            category,
            description: f.description || f.label,
          })
        )
      }
    }
    if (items.length === 0) return
    await batch.mutateAsync(new BatchUpdateSettingsRequest({ settings: items }))
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className='flex items-center gap-2 text-base'>
          <ReceiptText className='h-4 w-4 text-primary' /> {title}
        </CardTitle>
        <CardDescription>{description}</CardDescription>
      </CardHeader>
      <CardContent className='space-y-4'>
        {isLoading ? (
          <div className='flex items-center gap-2 text-sm text-muted-foreground'>
            <Loader2 className='h-4 w-4 animate-spin' /> Memuat pengaturan…
          </div>
        ) : (
          <div className='grid gap-4 md:grid-cols-2'>
            {fields.map((f) => (
              <div key={f.key} className='space-y-1.5'>
                <Label>{f.label}</Label>
                <FieldInput
                  def={f}
                  value={valueOf(f.key)}
                  onChange={(value) => setOverrides((prev) => ({ ...prev, [f.key]: value }))}
                />
                {f.description && (
                  <p className='text-xs text-muted-foreground'>{f.description}</p>
                )}
              </div>
            ))}
          </div>
        )}
        <div className='flex justify-end'>
          <Button onClick={handleSave} disabled={!dirty || batch.isPending} className='gap-1.5'>
            <Save className='h-4 w-4' />
            {batch.isPending ? 'Menyimpan…' : 'Simpan'}
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}

export function BillingSettingsForm() {
  return (
    <div className='space-y-6'>
      <CategoryCard
        title='Siklus & Tagihan'
        description='Mode billing, tutup buku, PPN default, pengingat, dan pembayaran massal. Tanpa denda dan tanpa biaya tambahan.'
        category='isp_billing'
        fields={BILLING_FIELDS}
      />
      <CategoryCard
        title='Operasional Gateway'
        description='Sinkronisasi transaksi PENDING dan kategori kas untuk biaya gateway.'
        category='isp_gateway'
        fields={GATEWAY_OPS_FIELDS}
      />
    </div>
  )
}
