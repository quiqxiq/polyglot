import { Link } from '@tanstack/react-router'
import { ArrowRightIcon, ShieldCheckIcon } from 'lucide-react'
import { Button } from '@/components/ui/button'
import type { AiOpsApproval } from '../api/ag-ui-mapping'

/**
 * A proposed mutation always lands in the Approval Inbox. The chat never
 * approves anything itself — that stays a human decision behind
 * `approval:decide`.
 */
export function ChatApprovalCard({ approval }: { approval: AiOpsApproval }) {
  return (
    <div className='flex flex-wrap items-center justify-between gap-2 rounded-md border border-amber-500/30 bg-amber-500/10 px-3 py-2'>
      <div className='flex min-w-0 items-center gap-2'>
        <ShieldCheckIcon className='size-4 shrink-0 text-amber-600 dark:text-amber-400' />
        <div className='min-w-0'>
          <p className='text-xs font-medium text-foreground'>
            Proposal menunggu persetujuan
          </p>
          <p className='truncate text-[11px] text-muted-foreground'>
            {approval.summary || approval.approvalId}
          </p>
        </div>
      </div>
      <Button asChild size='sm' className='h-7 gap-1 text-xs'>
        <Link to='/ai-ops/approvals'>
          Review Proposal <ArrowRightIcon className='size-3' />
        </Link>
      </Button>
    </div>
  )
}
