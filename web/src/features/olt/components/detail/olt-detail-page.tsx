import { useMemo, useState } from 'react'
import { useNavigate, useParams } from '@tanstack/react-router'
import type { Onu } from '@/gen/v1/olt_pb'
import { ChevronLeft, RefreshCw, Save } from 'lucide-react'
import { toast } from 'sonner'
import { useI18n } from '@/context/i18n-provider'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { ConfigDrawer } from '@/components/config-drawer'
import { ConfirmDialog } from '@/components/confirm-dialog'
import { LanguageSwitch } from '@/components/language-switch'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import {
  useDeleteOnuMutation,
  useOltQuery,
  useOnusQuery,
  useRebootOnuMutation,
  useSaveOltConfigMutation,
  useSetOnuAdminStateMutation,
  useSyncOltMutation,
} from '../../api/use-olt'
import { formatLatency, positionLabel } from '../../lib/format'
import { EnabledBadge, HealthBadge } from '../olt-badges'
import { OltJobButton } from '../olt-job-button'
import { ALL_FILTER, OltOnuFilters } from './olt-onu-filters'
import { OltOnuTable } from './olt-onu-table'
import { OnuDetailSheet } from './onu-detail-sheet'
import { OnuRenameDialog } from './onu-rename-dialog'

export function OltDetailPage() {
  const { t } = useI18n()
  const navigate = useNavigate()
  const { oltId } = useParams({ from: '/_authenticated/olt/$oltId' })

  const {
    data: detail,
    isLoading,
    isError,
    isFetching,
    refetch,
  } = useOltQuery(oltId)
  const olt = detail?.olt
  const health = detail?.health

  const [ponFilter, setPonFilter] = useState(ALL_FILTER)
  const [stateFilter, setStateFilter] = useState(ALL_FILTER)
  const [search, setSearch] = useState('')

  const { data: onus = [], isLoading: isLoadingOnus } = useOnusQuery(
    oltId,
    ponFilter === ALL_FILTER ? 0 : Number(ponFilter),
    stateFilter === ALL_FILTER ? '' : stateFilter
  )

  const [selectedOnu, setSelectedOnu] = useState<Onu | null>(null)
  const [renaming, setRenaming] = useState<Onu | null>(null)
  const [rebooting, setRebooting] = useState<Onu | null>(null)
  const [togglingAdmin, setTogglingAdmin] = useState<Onu | null>(null)
  const [deleting, setDeleting] = useState<Onu | null>(null)

  const syncMutation = useSyncOltMutation()
  const saveConfigMutation = useSaveOltConfigMutation()
  const rebootMutation = useRebootOnuMutation()
  const adminMutation = useSetOnuAdminStateMutation()
  const deleteOnuMutation = useDeleteOnuMutation()

  const filteredOnus = useMemo(() => {
    const query = search.trim().toLowerCase()
    if (!query) return onus
    return onus.filter(
      (onu) =>
        onu.name.toLowerCase().includes(query) ||
        onu.macAddress.toLowerCase().includes(query) ||
        onu.serialNumber.toLowerCase().includes(query)
    )
  }, [onus, search])

  const handleReboot = async () => {
    if (!rebooting) return
    try {
      const res = await rebootMutation.mutateAsync(rebooting)
      toast.success(t('olt.job_queued', { jobId: res.jobId || '-' }))
      setRebooting(null)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t('olt.job_failed'))
    }
  }

  const handleAdmin = async () => {
    if (!togglingAdmin) return
    const enabled = togglingAdmin.adminState !== 'ENABLE'
    try {
      const res = await adminMutation.mutateAsync({
        onu: togglingAdmin,
        enabled,
      })
      toast.success(t('olt.job_queued', { jobId: res.jobId || '-' }))
      setTogglingAdmin(null)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t('olt.job_failed'))
    }
  }

  const handleDelete = async () => {
    if (!deleting) return
    try {
      const res = await deleteOnuMutation.mutateAsync(deleting)
      toast.success(t('olt.job_queued', { jobId: res.jobId || '-' }))
      setDeleting(null)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t('olt.job_failed'))
    }
  }

  if (isLoading) {
    return (
      <>
        <Header fixed>
          <Search className='me-auto' />
          <LanguageSwitch />
          <ThemeSwitch />
          <ConfigDrawer />
          <ProfileDropdown />
        </Header>
        <Main className='flex flex-1 flex-col gap-4'>
          <Skeleton className='h-8 w-64' />
          <Skeleton className='h-24 w-full' />
          <Skeleton className='h-64 w-full' />
        </Main>
      </>
    )
  }

  if (isError || !olt) {
    return (
      <>
        <Header fixed>
          <Search className='me-auto' />
          <LanguageSwitch />
          <ThemeSwitch />
          <ConfigDrawer />
          <ProfileDropdown />
        </Header>
        <Main className='flex flex-1 flex-col gap-4'>
          <p className='text-sm text-muted-foreground'>{t('alerts.failed')}</p>
          <Button
            variant='outline'
            className='w-fit'
            onClick={() => navigate({ to: '/olt' })}
          >
            <ChevronLeft className='mr-1.5 h-4 w-4' />
            {t('olt.back')}
          </Button>
        </Main>
      </>
    )
  }

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
        <Button
          variant='ghost'
          size='sm'
          className='w-fit px-0 text-muted-foreground'
          onClick={() => navigate({ to: '/olt' })}
        >
          <ChevronLeft className='h-4 w-4' />
          {t('olt.back')}
        </Button>

        <div className='flex flex-wrap items-start justify-between gap-3'>
          <div>
            <div className='flex flex-wrap items-center gap-2'>
              <h1 className='text-2xl font-bold tracking-tight'>
                {olt.name || olt.host}
              </h1>
              <Badge variant='outline'>{olt.variant || 'AUTO'}</Badge>
              <EnabledBadge enabled={olt.enabled} />
              <HealthBadge health={health} />
            </div>
            <p className='text-sm text-muted-foreground'>
              {olt.host} · {olt.transport.toUpperCase()} · {t('olt.pon_count')}:{' '}
              {olt.ponCount}
            </p>
          </div>
          <div className='flex flex-wrap items-center gap-2'>
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
              {t('olt.refresh_health')}
            </Button>
            <OltJobButton
              label={t('olt.sync')}
              icon={RefreshCw}
              size='sm'
              className='h-9'
              onRun={async () => {
                const res = await syncMutation.mutateAsync(oltId)
                return res.jobId
              }}
              errorMessage={t('olt.sync_failed')}
            />
            <OltJobButton
              label={t('olt.save_config')}
              icon={Save}
              size='sm'
              className='h-9'
              onRun={async () => {
                const res = await saveConfigMutation.mutateAsync(oltId)
                return res.jobId
              }}
              errorMessage={t('olt.save_config_failed')}
            />
          </div>
        </div>

        <div className='grid gap-3 rounded-xl border p-4 sm:grid-cols-2 lg:grid-cols-4'>
          <div className='text-sm'>
            <span className='text-muted-foreground'>{t('olt.uptime')}</span>
            <p className='font-semibold'>{health?.uptime || '-'}</p>
          </div>
          <div className='text-sm'>
            <span className='text-muted-foreground'>{t('olt.firmware')}</span>
            <p className='font-semibold'>{health?.firmware || '-'}</p>
          </div>
          <div className='text-sm'>
            <span className='text-muted-foreground'>{t('olt.latency')}</span>
            <p className='font-semibold'>{formatLatency(health?.latencyMs)}</p>
          </div>
          <div className='text-sm'>
            <span className='text-muted-foreground'>{t('olt.sys_descr')}</span>
            <p className='line-clamp-2 font-mono text-xs'>
              {health?.sysDescr || '-'}
            </p>
          </div>
        </div>

        <div className='grid grid-cols-2 gap-3 lg:grid-cols-4'>
          {[
            { label: t('olt.total_onus'), value: olt.totalOnus },
            { label: t('olt.online_onus'), value: olt.onlineOnus },
            { label: t('olt.offline_onus'), value: olt.offlineOnus },
            { label: t('olt.unknown_onus'), value: olt.unknownOnus },
          ].map((item) => (
            <div key={item.label} className='rounded-lg border p-3'>
              <p className='text-xs text-muted-foreground'>{item.label}</p>
              <p className='text-xl font-bold'>{item.value}</p>
            </div>
          ))}
        </div>

        <OltOnuFilters
          ponCount={olt.ponCount}
          ponIndex={ponFilter}
          state={stateFilter}
          search={search}
          isFetching={isFetching}
          onPonChange={setPonFilter}
          onStateChange={setStateFilter}
          onSearchChange={setSearch}
          onRefresh={() => refetch()}
        />

        <OltOnuTable
          onus={filteredOnus}
          isLoading={isLoadingOnus}
          onSelect={setSelectedOnu}
          onRename={setRenaming}
          onReboot={setRebooting}
          onToggleAdmin={setTogglingAdmin}
          onDelete={setDeleting}
        />
      </Main>

      <OnuDetailSheet
        key={selectedOnu?.id ?? 'empty'}
        onuId={selectedOnu?.id ?? null}
        onOpenChange={(open) => {
          if (!open) setSelectedOnu(null)
        }}
      />

      {renaming && (
        <OnuRenameDialog onu={renaming} onClose={() => setRenaming(null)} />
      )}

      <ConfirmDialog
        open={rebooting !== null}
        onOpenChange={(open) => {
          if (!open) setRebooting(null)
        }}
        title={t('olt.reboot_title')}
        desc={
          rebooting
            ? t('olt.reboot_confirm', {
                position: positionLabel(rebooting.ponIndex, rebooting.onuIndex),
              })
            : ''
        }
        confirmText={t('olt.reboot')}
        isLoading={rebootMutation.isPending}
        handleConfirm={handleReboot}
      />

      <ConfirmDialog
        open={togglingAdmin !== null}
        onOpenChange={(open) => {
          if (!open) setTogglingAdmin(null)
        }}
        title={t('olt.admin_title')}
        desc={
          togglingAdmin
            ? togglingAdmin.adminState === 'ENABLE'
              ? t('olt.admin_confirm_disable', {
                  position: positionLabel(
                    togglingAdmin.ponIndex,
                    togglingAdmin.onuIndex
                  ),
                })
              : t('olt.admin_confirm_enable', {
                  position: positionLabel(
                    togglingAdmin.ponIndex,
                    togglingAdmin.onuIndex
                  ),
                })
            : ''
        }
        confirmText={t('common.confirm')}
        isLoading={adminMutation.isPending}
        handleConfirm={handleAdmin}
      />

      <ConfirmDialog
        open={deleting !== null}
        onOpenChange={(open) => {
          if (!open) setDeleting(null)
        }}
        title={t('olt.delete_onu_title')}
        desc={
          deleting
            ? t('olt.delete_onu_confirm', {
                position: positionLabel(deleting.ponIndex, deleting.onuIndex),
              })
            : ''
        }
        confirmText={t('olt.delete_onu')}
        destructive
        isLoading={deleteOnuMutation.isPending}
        handleConfirm={handleDelete}
      />
    </>
  )
}
