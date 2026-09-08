import { useState } from 'react'
import { Link } from '@tanstack/react-router'
import { Server, FileSpreadsheet, ArrowRight, Sparkles } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { useDevicesQuery } from '@/features/devices/api/use-devices'
import { CustomersImportDialogFileTab } from './customers-import-dialog-file-tab'
import { CustomersImportDialogRouterTab } from './customers-import-dialog-router-tab'

type CustomersImportDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function CustomersImportDialog({
  open,
  onOpenChange,
}: CustomersImportDialogProps) {
  const [activeTab, setActiveTab] = useState<'router' | 'file'>('file')
  const { data: devices = [], isLoading: devicesLoading } = useDevicesQuery()

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='max-h-[90vh] gap-4 overflow-y-auto sm:max-w-2xl'>
        <DialogHeader className='text-start'>
          <div className='flex items-center gap-2'>
            <div className='flex h-9 w-9 items-center justify-center rounded-lg bg-primary/10 text-primary'>
              <Server className='h-5 w-5' />
            </div>
            <div>
              <DialogTitle className='text-lg'>Import Pelanggan</DialogTitle>
              <DialogDescription>
                Hubungkan data pelanggan dan langganan aktif ke Router MikroTik
                fisik.
              </DialogDescription>
            </div>
          </div>
        </DialogHeader>

        {/* Banner Link to Full Page Wizard */}
        <div className='flex flex-wrap items-center justify-between gap-2 rounded-lg border border-primary/30 bg-primary/5 p-3 text-xs'>
          <div className='flex items-center gap-2'>
            <Sparkles className='h-4 w-4 text-primary' />
            <span className='font-medium text-foreground'>
              Ingin mengedit nama, no HP, atau paket sebelum disimpan?
            </span>
          </div>
          <Button size='sm' variant='outline' className='h-7 text-xs' asChild>
            <Link to='/customers/import' onClick={() => onOpenChange(false)}>
              Buka Halaman Penuh <ArrowRight className='ml-1 h-3 w-3' />
            </Link>
          </Button>
        </div>

        <Tabs
          value={activeTab}
          onValueChange={(val) => setActiveTab(val as 'router' | 'file')}
          className='w-full'
        >
          <TabsList className='grid w-full grid-cols-2'>
            <TabsTrigger value='router' className='gap-2'>
              <Server className='h-4 w-4' />
              <span>Metode A: Tarik dari Router</span>
            </TabsTrigger>
            <TabsTrigger value='file' className='gap-2'>
              <FileSpreadsheet className='h-4 w-4' />
              <span>Metode B: Upload Excel / CSV</span>
            </TabsTrigger>
          </TabsList>

          {/* Tab 1: Metode A (Router Live Pull) */}
          <TabsContent value='router'>
            <CustomersImportDialogRouterTab
              devices={devices}
              devicesLoading={devicesLoading}
              onSuccess={() => onOpenChange(false)}
            />
          </TabsContent>

          {/* Tab 2: Metode B (Upload File) */}
          <TabsContent value='file'>
            <CustomersImportDialogFileTab
              devices={devices}
              devicesLoading={devicesLoading}
              onSuccess={() => onOpenChange(false)}
            />
          </TabsContent>
        </Tabs>
      </DialogContent>
    </Dialog>
  )
}
