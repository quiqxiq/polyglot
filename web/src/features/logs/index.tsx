import { useState, useMemo, useCallback } from 'react'
import { ScrollText, AlertCircle } from 'lucide-react'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { ConfigDrawer } from '@/components/config-drawer'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { useDeviceStore } from '@/stores/device-store'
import { useDevicesQuery } from '@/features/devices/api/use-devices'
import { useLogStream } from './api/use-log-stream'
import { LogToolbar } from './components/log-toolbar'
import { LogsTable } from './components/logs-table'
import { StructuredLogTable } from './components/structured-log-table'
import { LogItem, SeverityFilter } from './types'
import { exportLogsToJson, exportLogsToTxt } from './lib/log-formatter'

export function LogsFeature() {
  const { selectedDeviceId } = useDeviceStore()
  const { data: devices = [] } = useDevicesQuery()
  const currentDevice = devices.find((d) => d.id === selectedDeviceId)

  // Search & Filter state
  const [searchTerm, setSearchTerm] = useState('')
  const [severityFilter, setSeverityFilter] = useState<SeverityFilter>('all')
  const [isAutoScroll, setIsAutoScroll] = useState(true)

  // Persist viewMode in localStorage so it never unexpectedly switches back to default
  const [viewMode, setViewModeState] = useState<'structured' | 'raw'>(() => {
    if (typeof window !== 'undefined') {
      const saved = localStorage.getItem('logs_view_mode')
      if (saved === 'raw' || saved === 'structured') return saved
    }
    return 'structured'
  })

  const setViewMode = useCallback((mode: 'structured' | 'raw') => {
    setViewModeState(mode)
    if (typeof window !== 'undefined') {
      localStorage.setItem('logs_view_mode', mode)
    }
  }, [])

  // RouterOS live log stream
  const stream = useLogStream({
    deviceId: selectedDeviceId || '',
    topics: '',
    enabled: Boolean(selectedDeviceId),
  })

  // Filter logs by search term & severity on the client side
  const filteredLogs = useMemo(() => {
    return stream.logs.filter((item: LogItem) => {
      // 1. Severity filter
      if (severityFilter !== 'all' && item.severity !== severityFilter) {
        return false
      }

      // 2. Search keyword filter
      if (!searchTerm || !searchTerm.trim()) {
        return true
      }

      const q = searchTerm.toLowerCase().trim()
      return (
        item.message.toLowerCase().includes(q) ||
        item.topics.toLowerCase().includes(q) ||
        item.time.toLowerCase().includes(q)
      )
    })
  }, [stream.logs, severityFilter, searchTerm])

  const handleExportTxt = () => {
    const filename = `${currentDevice?.name || 'router'}-all-logs.txt`
    exportLogsToTxt(filteredLogs, filename)
  }

  const handleExportJson = () => {
    const filename = `${currentDevice?.name || 'router'}-all-logs.json`
    exportLogsToJson(filteredLogs, filename)
  }

  return (
    <>
      <Header fixed>
        <Search className='me-auto' />
        <ThemeSwitch />
        <ConfigDrawer />
        <ProfileDropdown />
      </Header>

      <Main fixed className='flex flex-1 flex-col gap-3 min-h-0 overflow-hidden'>
        {/* ===== Title & Router Header (Fixed / Non-shrinking) ===== */}
        <div className='shrink-0 flex flex-wrap items-end justify-between gap-2'>
          <div>
            <div className='flex items-center gap-2'>
              <ScrollText className='size-5 text-primary sm:size-6' />
              <h1 className='text-xl font-bold tracking-tight sm:text-2xl'>Router Live Logs</h1>
            </div>
            <p className='text-muted-foreground mt-0.5 text-xs sm:text-sm'>
              {currentDevice ? (
                <>
                  Router:{' '}
                  <span className='text-foreground font-semibold'>
                    {currentDevice.name}
                  </span>{' '}
                  ({currentDevice.host}) &bull; Vendor:{' '}
                  <span className='capitalize font-medium text-foreground'>
                    {currentDevice.vendor}
                  </span>
                </>
              ) : (
                'Select a MikroTik router from the sidebar to stream live system events.'
              )}
            </p>
          </div>
        </div>

        {/* If no router selected */}
        {!selectedDeviceId ? (
          <div className='flex flex-1 flex-col items-center justify-center rounded-lg border border-dashed p-12 text-center'>
            <AlertCircle className='text-muted-foreground mb-3 size-10' />
            <h3 className='text-lg font-semibold'>No Router Selected</h3>
            <p className='text-muted-foreground mt-1 max-w-sm text-sm'>
              Please select a MikroTik router from the top selector in the sidebar to begin streaming live logs.
            </p>
          </div>
        ) : (
          <div className='flex flex-1 min-h-0 flex-col gap-2.5 overflow-hidden'>
            {/* Controls & Filter Toolbar (Fixed) */}
            <div className='shrink-0'>
              <LogToolbar
                searchTerm={searchTerm}
                onSearchChange={setSearchTerm}
                severityFilter={severityFilter}
                onSeverityChange={setSeverityFilter}
                isPaused={stream.isPaused}
                onTogglePause={stream.togglePause}
                isAutoScroll={isAutoScroll}
                onToggleAutoScroll={() => setIsAutoScroll((prev) => !prev)}
                onClear={stream.clearLogs}
                onExportTxt={handleExportTxt}
                onExportJson={handleExportJson}
                totalCount={stream.logs.length}
                filteredCount={filteredLogs.length}
                isStreaming={stream.isStreaming}
              />
            </div>

            {/* Main Log Viewer: Structured Table View or Raw Stream View */}
            <div className='flex flex-1 min-h-0 flex-col overflow-hidden'>
              {viewMode === 'structured' ? (
                <StructuredLogTable
                  logs={filteredLogs}
                  highlight={searchTerm}
                  isAutoScroll={isAutoScroll}
                  isLoading={stream.isLoading}
                  viewMode={viewMode}
                  onViewModeChange={setViewMode}
                />
              ) : (
                <LogsTable
                  logs={filteredLogs}
                  highlight={searchTerm}
                  isAutoScroll={isAutoScroll}
                  isLoading={stream.isLoading}
                  viewMode={viewMode}
                  onViewModeChange={setViewMode}
                />
              )}
            </div>
          </div>
        )}
      </Main>
    </>
  )
}
