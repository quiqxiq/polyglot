import { useMemo, useState } from 'react'
import { useNavigate } from '@tanstack/react-router'
import type { OltDevice, TestOltConnectionResponse } from '@/gen/v1/olt_pb'
import { Plus, RefreshCw, SearchIcon } from 'lucide-react'
import { toast } from 'sonner'
import { useI18n } from '@/context/i18n-provider'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { ConfigDrawer } from '@/components/config-drawer'
import { LanguageSwitch } from '@/components/language-switch'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import {
  useOltsQuery,
  useSyncOltMutation,
  useTestOltConnectionMutation,
} from './api/use-olt'
import { OltDeleteDialog } from './components/olt-delete-dialog'
import { OltFormDialog } from './components/olt-form-dialog'
import { OltTable } from './components/olt-table'
import { OltTestDialog } from './components/olt-test-dialog'

/**
 * Halaman inventori OLT EPON HiOSO: daftar perangkat, CRUD, test koneksi,
 * dan trigger sync ONU via job queue.
 */
export function OltListPage() {
  const { t } = useI18n()
  const navigate = useNavigate()
  const { data: olts = [], isLoading, refetch, isFetching } = useOltsQuery()
  const testMutation = useTestOltConnectionMutation()
  const syncMutation = useSyncOltMutation()

  const [search, setSearch] = useState('')
  const [formOpen, setFormOpen] = useState(false)
  const [editing, setEditing] = useState<OltDevice | null>(null)
  const [deleting, setDeleting] = useState<OltDevice | null>(null)
  const [testTarget, setTestTarget] = useState<OltDevice | null>(null)
  const [testResult, setTestResult] =
    useState<TestOltConnectionResponse | null>(null)
  const [testError, setTestError] = useState<string | null>(null)

  const filtered = useMemo(() => {
    const query = search.trim().toLowerCase()
    if (!query) return olts
    return olts.filter(
      (olt) =>
        olt.name.toLowerCase().includes(query) ||
        olt.host.toLowerCase().includes(query)
    )
  }, [olts, search])

  const openCreate = () => {
    setEditing(null)
    setFormOpen(true)
  }

  const openEdit = (olt: OltDevice) => {
    setEditing(olt)
    setFormOpen(true)
  }

  const openDetail = (olt: OltDevice) => {
    navigate({ to: '/olt/$oltId', params: { oltId: olt.id } })
  }

  const handleTest = async (olt: OltDevice) => {
    setTestTarget(olt)
    setTestResult(null)
    setTestError(null)
    try {
      const res = await testMutation.mutateAsync(olt.id)
      setTestResult(res)
    } catch (err) {
      setTestError(err instanceof Error ? err.message : t('olt.test_failed'))
    }
  }

  const handleSync = async (olt: OltDevice) => {
    try {
      const res = await syncMutation.mutateAsync(olt.id)
      toast.success(t('olt.job_queued', { jobId: res.jobId || '-' }))
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t('olt.sync_failed'))
    }
  }

  const onlineOnus = olts.reduce((sum, olt) => sum + olt.onlineOnus, 0)
  const totalOnus = olts.reduce((sum, olt) => sum + olt.totalOnus, 0)

  return (
    <>
      <Header fixed>
        <Search className='me-auto' />
        <LanguageSwitch />
        <ThemeSwitch />
        <ConfigDrawer />
        <ProfileDropdown />
      </Header>

      <Main className='flex flex-1 flex-col gap-4 sm:gap-6'>
        <div className='flex flex-wrap items-end justify-between gap-2'>
          <div>
            <h1 className='text-2xl font-bold tracking-tight'>
              {t('olt.title')}
            </h1>
            <p className='text-sm text-muted-foreground'>{t('olt.subtitle')}</p>
          </div>
          <Button onClick={openCreate} className='h-9'>
            <Plus className='mr-1.5 h-4 w-4' />
            {t('olt.add')}
          </Button>
        </div>

        <div className='flex flex-wrap items-center gap-2 text-xs text-muted-foreground'>
          <span className='rounded-md border px-2 py-1'>{olts.length} OLT</span>
          <span className='rounded-md border border-emerald-500/30 bg-emerald-500/10 px-2 py-1 text-emerald-700 dark:text-emerald-400'>
            {onlineOnus} / {totalOnus} {t('olt.online_onus')}
          </span>
        </div>

        <div className='flex flex-wrap items-center gap-2'>
          <div className='relative w-full sm:max-w-xs'>
            <SearchIcon className='absolute start-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground' />
            <Input
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder={t('olt.search_placeholder')}
              className='ps-8'
            />
          </div>
          <Button
            variant='outline'
            size='sm'
            className='h-9'
            onClick={() => refetch()}
            disabled={isFetching}
          >
            <RefreshCw
              className={`mr-1.5 h-4 w-4 ${isFetching ? 'animate-spin' : ''}`}
            />
            {t('common.refresh')}
          </Button>
        </div>

        <OltTable
          olts={filtered}
          isLoading={isLoading}
          onDetail={openDetail}
          onEdit={openEdit}
          onDelete={setDeleting}
          onTest={handleTest}
          onSync={handleSync}
        />
      </Main>

      {formOpen && (
        <OltFormDialog
          olt={editing}
          onClose={() => {
            setFormOpen(false)
            setEditing(null)
          }}
        />
      )}

      {deleting && (
        <OltDeleteDialog olt={deleting} onClose={() => setDeleting(null)} />
      )}

      <OltTestDialog
        open={testTarget !== null}
        olt={testTarget}
        result={testResult}
        isPending={testMutation.isPending}
        error={testError}
        onOpenChange={(open) => {
          if (!open) {
            setTestTarget(null)
            setTestResult(null)
            setTestError(null)
          }
        }}
        onRetry={() => {
          if (testTarget) void handleTest(testTarget)
        }}
      />
    </>
  )
}
