import { useState } from 'react'
import { toast } from 'sonner'
import { Undo2 } from 'lucide-react'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { Button } from '@/components/ui/button'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import type { ManualPaymentSubmission } from '@/gen/v1/billing_pb'
import {
  useApprovePaymentSubmissionMutation,
  usePaymentSubmissionsQuery,
} from './api/use-payments'
import { PaymentsTable } from './components/payments-table'
import { RejectPaymentDialog } from './components/reject-payment-dialog'
import { ReversePaymentDialog } from './components/reverse-payment-dialog'

const STATUS_TABS = [
  { value: 'PENDING', label: 'Menunggu Review' },
  { value: 'APPROVED', label: 'Disetujui' },
  { value: 'REJECTED', label: 'Ditolak' },
  { value: '', label: 'Semua' },
]

/**
 * Halaman antrean approval bukti transfer manual + reversal pembayaran
 * (planning v2 Fase 3 — "Operasional Pembayaran v2").
 */
export function Payments() {
  const [status, setStatus] = useState('PENDING')
  const [rejectTarget, setRejectTarget] =
    useState<ManualPaymentSubmission | null>(null)
  const [reverseOpen, setReverseOpen] = useState(false)
  const { data = [], isLoading } = usePaymentSubmissionsQuery(status)
  const approveMutation = useApprovePaymentSubmissionMutation()

  const handleApprove = async (sub: ManualPaymentSubmission) => {
    try {
      await approveMutation.mutateAsync({ id: sub.id })
      toast.success('Bukti transfer disetujui; invoice lunas & router dipulihkan.')
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Gagal menyetujui bukti transfer')
    }
  }

  return (
    <>
      <Header fixed>
        <Search className='me-auto' />
        <ThemeSwitch />
        <ProfileDropdown />
      </Header>
      <Main className='flex flex-1 flex-col gap-4 sm:gap-6'>
        <div className='flex flex-wrap items-end justify-between gap-2'>
          <div>
            <h2 className='text-2xl font-bold tracking-tight'>
              Pembayaran Manual & Reversal
            </h2>
            <p className='text-muted-foreground'>
              Antrean approval bukti transfer pelanggan dan koreksi pembayaran
              yang salah input.
            </p>
          </div>
          <Button variant='outline' onClick={() => setReverseOpen(true)}>
            <Undo2 className='mr-1.5 h-4 w-4' /> Reversal Pembayaran
          </Button>
        </div>

        <Tabs value={status} onValueChange={setStatus}>
          <TabsList className='flex h-9'>
            {STATUS_TABS.map((tab) => (
              <TabsTrigger key={tab.value} value={tab.value} className='text-xs'>
                {tab.label}
              </TabsTrigger>
            ))}
          </TabsList>
        </Tabs>

        <PaymentsTable
          data={data}
          isLoading={isLoading}
          onApprove={handleApprove}
          onReject={(sub) => setRejectTarget(sub)}
        />
      </Main>

      <RejectPaymentDialog
        open={rejectTarget !== null}
        onOpenChange={(open) => !open && setRejectTarget(null)}
        submission={rejectTarget}
      />
      <ReversePaymentDialog open={reverseOpen} onOpenChange={setReverseOpen} />
    </>
  )
}
