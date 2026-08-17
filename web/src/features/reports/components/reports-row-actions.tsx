import { Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import type { HotspotReport } from '@/gen/v1/hotspot_pb'
import { useReports } from '../context/reports-context'

interface ReportsRowActionsProps {
  report: HotspotReport
}

export function ReportsRowActions({ report }: ReportsRowActionsProps) {
  const { setOpen, setCurrentReport } = useReports()

  const handleDelete = () => {
    setCurrentReport(report)
    setOpen('report-delete')
  }

  return (
    <Button
      variant='ghost'
      size='sm'
      onClick={handleDelete}
      className='h-8 px-2 text-destructive hover:text-destructive hover:bg-destructive/10'
      title='Delete transaction'
    >
      <Trash2 className='size-3.5 mr-1' />
      Delete
    </Button>
  )
}
