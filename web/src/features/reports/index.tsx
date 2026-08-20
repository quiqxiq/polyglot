import { useState } from 'react'
import { FileText, AlertCircle, RefreshCw } from 'lucide-react'
import { ConfigDrawer } from '@/components/config-drawer'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { ReportsProvider } from './context/reports-context'
import { ReportsSummaryCards } from './components/reports-summary-cards'
import { ReportsTable } from './components/reports-table'
import { ReportDeleteDialog } from './components/report-delete-dialog'
import { useHotspotReportsQuery } from './api/use-reports'
import { useDeviceStore } from '@/stores/device-store'
import { useDevicesQuery } from '@/features/devices/api/use-devices'

function ReportsContent() {
  const { selectedDeviceId } = useDeviceStore()
  const { data: devices = [] } = useDevicesQuery()

  const [dayFilter, setDayFilter] = useState('')
  const [monthFilter, setMonthFilter] = useState('')

  const { data, isLoading, refetch } = useHotspotReportsQuery(
    selectedDeviceId,
    dayFilter,
    monthFilter
  )

  const currentDevice = devices.find((d) => d.id === selectedDeviceId)
  const reports = data?.reports || []
  const totalIncome = data?.totalIncome || 0
  const total = data?.total || reports.length

  const filterLabel = dayFilter
    ? `Day: ${dayFilter}`
    : monthFilter
      ? `Month: ${monthFilter}`
      : 'All Recorded Transactions'

  return (
    <>
      <Header fixed>
        <Search className='me-auto' />
        <ThemeSwitch />
        <ConfigDrawer />
        <ProfileDropdown />
      </Header>

      <Main className='flex flex-1 flex-col gap-4 sm:gap-6'>
        {/* Title */}
        <div className='flex flex-wrap items-end justify-between gap-2'>
          <div>
            <div className='flex items-center gap-2'>
              <FileText className='size-6 text-primary' />
              <h1 className='text-2xl font-bold tracking-tight'>Sales Report</h1>
            </div>
            <p className='text-sm text-muted-foreground mt-0.5'>
              {currentDevice ? (
                <>
                  Router:{' '}
                  <span className='font-semibold text-foreground'>
                    {currentDevice.name}
                  </span>{' '}
                  ({currentDevice.host})
                </>
              ) : (
                'Select a router in the sidebar to view sales records.'
              )}
            </p>
          </div>

          <div className='flex items-center gap-2'>
            <Button
              variant='outline'
              size='sm'
              onClick={() => refetch()}
              disabled={!selectedDeviceId || isLoading}
              className='gap-1.5 h-9'
            >
              <RefreshCw className={`size-3.5 ${isLoading ? 'animate-spin' : ''}`} />
              Refresh
            </Button>
          </div>
        </div>

        {!selectedDeviceId ? (
          <div className='flex flex-col items-center justify-center p-12 text-center rounded-lg border border-dashed'>
            <AlertCircle className='size-10 text-muted-foreground mb-3' />
            <h3 className='text-lg font-semibold'>No Router Selected</h3>
            <p className='text-sm text-muted-foreground max-w-sm mt-1'>
              Please select a MikroTik router from the top dropdown in the sidebar to view sales and income reports.
            </p>
          </div>
        ) : (
          <>
            {/* Quick Date Filters Toolbar */}
            <div className='flex flex-wrap items-center gap-2 bg-card p-3 rounded-lg border'>
              <span className='text-xs font-semibold text-muted-foreground mr-1'>Filter Date:</span>
              <Input
                placeholder='Day (e.g. aug/17/2026)'
                className='h-8 w-48 text-xs font-mono'
                value={dayFilter}
                onChange={(e) => {
                  setDayFilter(e.target.value)
                  if (e.target.value) setMonthFilter('')
                }}
              />
              <Input
                placeholder='Month (e.g. aug2026)'
                className='h-8 w-44 text-xs font-mono'
                value={monthFilter}
                onChange={(e) => {
                  setMonthFilter(e.target.value)
                  if (e.target.value) setDayFilter('')
                }}
              />
              {(dayFilter || monthFilter) && (
                <Button
                  variant='ghost'
                  size='sm'
                  className='h-8 text-xs'
                  onClick={() => {
                    setDayFilter('')
                    setMonthFilter('')
                  }}
                >
                  Clear Filter
                </Button>
              )}
            </div>

            {/* Income Summary Cards */}
            <ReportsSummaryCards
              totalIncome={totalIncome}
              totalCount={total}
              filterLabel={filterLabel}
            />

            {/* Sales Table */}
            <ReportsTable data={reports} isLoading={isLoading} />
          </>
        )}
      </Main>

      <ReportDeleteDialog />
    </>
  )
}

export function Reports() {
  return (
    <ReportsProvider>
      <ReportsContent />
    </ReportsProvider>
  )
}
