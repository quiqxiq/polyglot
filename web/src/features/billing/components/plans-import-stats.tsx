import type { PlanImportRow } from '@/gen/v1/ispadmin_pb'
import { Layers, Sparkles, CheckSquare, Zap, Radio } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

interface PlansImportStatsProps {
  rows: PlanImportRow[]
}

export function PlansImportStats({ rows }: PlansImportStatsProps) {
  const pppoeCount = rows.filter((r) => r.serviceType === 'PPPOE').length
  const hotspotCount = rows.filter((r) => r.serviceType === 'HOTSPOT').length
  const newCount = rows.filter((r) => r.isNew).length
  const selectedCount = rows.filter((r) => r.selected).length

  return (
    <div className='grid grid-cols-2 gap-3 sm:grid-cols-5'>
      {/* 1. Total Terdeteksi */}
      <Card className='shadow-xs'>
        <CardHeader className='flex flex-row items-center justify-between space-y-0 p-3 pb-1'>
          <CardTitle className='text-xs font-medium text-muted-foreground'>
            Total Profil
          </CardTitle>
          <Layers className='h-4 w-4 text-primary' />
        </CardHeader>
        <CardContent className='p-3 pt-0'>
          <div className='text-xl font-bold'>{rows.length}</div>
        </CardContent>
      </Card>

      {/* 2. PPPoE */}
      <Card className='shadow-xs'>
        <CardHeader className='flex flex-row items-center justify-between space-y-0 p-3 pb-1'>
          <CardTitle className='text-xs font-medium text-muted-foreground'>
            Profil PPPoE
          </CardTitle>
          <Zap className='h-4 w-4 text-sky-500' />
        </CardHeader>
        <CardContent className='p-3 pt-0'>
          <div className='text-xl font-bold text-sky-600 dark:text-sky-400'>
            {pppoeCount}
          </div>
        </CardContent>
      </Card>

      {/* 3. Hotspot */}
      <Card className='shadow-xs'>
        <CardHeader className='flex flex-row items-center justify-between space-y-0 p-3 pb-1'>
          <CardTitle className='text-xs font-medium text-muted-foreground'>
            Profil Hotspot
          </CardTitle>
          <Radio className='h-4 w-4 text-amber-500' />
        </CardHeader>
        <CardContent className='p-3 pt-0'>
          <div className='text-xl font-bold text-amber-600 dark:text-amber-400'>
            {hotspotCount}
          </div>
        </CardContent>
      </Card>

      {/* 4. Profil Baru */}
      <Card className='shadow-xs'>
        <CardHeader className='flex flex-row items-center justify-between space-y-0 p-3 pb-1'>
          <CardTitle className='text-xs font-medium text-muted-foreground'>
            Profil Baru
          </CardTitle>
          <Sparkles className='h-4 w-4 text-emerald-500' />
        </CardHeader>
        <CardContent className='p-3 pt-0'>
          <div className='text-xl font-bold text-emerald-600 dark:text-emerald-400'>
            {newCount}
          </div>
        </CardContent>
      </Card>

      {/* 5. Terpilih */}
      <Card className='shadow-xs'>
        <CardHeader className='flex flex-row items-center justify-between space-y-0 p-3 pb-1'>
          <CardTitle className='text-xs font-medium text-muted-foreground'>
            Dipilih
          </CardTitle>
          <CheckSquare className='h-4 w-4 text-primary' />
        </CardHeader>
        <CardContent className='p-3 pt-0'>
          <div className='text-xl font-bold text-primary'>{selectedCount}</div>
        </CardContent>
      </Card>
    </div>
  )
}
