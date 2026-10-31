import { useMemo, useState } from 'react'
import { ApprovalDecision, type Approval } from '@/gen/v1/approval_pb'
import { Check, Loader2, Pencil, X } from 'lucide-react'
import { toast } from 'sonner'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Textarea } from '@/components/ui/textarea'
import { useDecideApprovalMutation } from '../api/use-approvals'
import { DiffView } from './diff-view'

interface Props {
  approval: Approval
  open: boolean
  onOpenChange: (open: boolean) => void
  canDecide: boolean
}

function fmtTime(unix: number | bigint | undefined) {
  const n = Number(unix || 0)
  if (!n) return '-'
  return new Date(n * 1000).toLocaleString('id-ID')
}

function statusVariant(
  status: string
): 'default' | 'secondary' | 'destructive' | 'outline' {
  switch (status) {
    case 'PENDING':
      return 'default'
    case 'EXECUTED':
      return 'secondary'
    case 'FAILED':
    case 'REJECTED':
      return 'destructive'
    default:
      return 'outline'
  }
}

export function ApprovalDetailDialog({
  approval,
  open,
  onOpenChange,
  canDecide,
}: Props) {
  const decide = useDecideApprovalMutation()
  const [rejectOpen, setRejectOpen] = useState(false)
  const [reason, setReason] = useState('')
  const [editMode, setEditMode] = useState(false)
  const [editedCommand, setEditedCommand] = useState(approval.command)
  const [editedArgs, setEditedArgs] = useState(() =>
    JSON.stringify(approval.args, null, 2)
  )

  const pending = approval.status === 'PENDING'
  const revisionConflictText = useMemo(
    () => `revision ${approval.revision}`,
    [approval.revision]
  )

  async function decideWith(
    decision: ApprovalDecision,
    extra?: Partial<{ reason: string }>
  ) {
    try {
      await decide.mutateAsync({
        id: approval.id,
        decision,
        editedCommand: editMode ? editedCommand : '',
        editedArgs: editMode ? parseArgs(editedArgs) : {},
        reason: extra?.reason || '',
        revision: approval.revision,
      })
      setEditMode(false)
      onOpenChange(false)
    } catch {
      // error toast handled by hook
    }
  }

  function parseArgs(raw: string): Record<string, string> {
    try {
      const parsed = JSON.parse(raw) as Record<string, string>
      return parsed || {}
    } catch {
      toast.error('Args harus berupa JSON object')
      return {}
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='max-h-[90vh] overflow-y-auto sm:max-w-3xl'>
        <DialogHeader>
          <DialogTitle className='flex items-center gap-2'>
            Approval <span className='font-mono text-sm'>{approval.id}</span>
            <Badge variant={statusVariant(approval.status)}>
              {approval.status}
            </Badge>
          </DialogTitle>
          <DialogDescription>
            {approval.deviceId} · revisi {revisionConflictText} · dibuat{' '}
            {fmtTime(approval.createdAtUnix)} · kedaluwarsa{' '}
            {fmtTime(approval.expiresAtUnix)}
          </DialogDescription>
        </DialogHeader>

        <Tabs defaultValue='diff'>
          <TabsList className='flex h-auto flex-wrap justify-start'>
            <TabsTrigger value='diff'>Diff</TabsTrigger>
            <TabsTrigger value='command'>Command</TabsTrigger>
            <TabsTrigger value='validation'>Validation</TabsTrigger>
            <TabsTrigger value='restore'>Restore Point</TabsTrigger>
          </TabsList>
          <TabsContent value='diff' className='pt-2'>
            <DiffView raw={approval.diffJson} />
          </TabsContent>
          <TabsContent value='command' className='space-y-3 pt-2'>
            {editMode ? (
              <>
                <Textarea
                  className='font-mono text-xs'
                  rows={5}
                  value={editedCommand}
                  onChange={(e) => setEditedCommand(e.target.value)}
                />
                <Textarea
                  className='font-mono text-xs'
                  rows={4}
                  value={editedArgs}
                  onChange={(e) => setEditedArgs(e.target.value)}
                />
              </>
            ) : (
              <>
                <pre className='overflow-x-auto rounded-md border bg-muted/30 p-3 font-mono text-xs'>
                  {approval.command}
                </pre>
                {Object.keys(approval.args).length > 0 && (
                  <pre className='overflow-x-auto rounded-md border bg-muted/30 p-3 font-mono text-xs'>
                    {JSON.stringify(approval.args, null, 2)}
                  </pre>
                )}
                {approval.editedCommand && (
                  <p className='text-xs text-muted-foreground'>
                    Sebelumnya diedit:{' '}
                    <span className='font-mono'>{approval.editedCommand}</span>
                  </p>
                )}
              </>
            )}
          </TabsContent>
          <TabsContent value='validation' className='pt-2'>
            <pre className='overflow-x-auto rounded-md border bg-muted/30 p-3 font-mono text-xs'>
              {approval.validationJson ||
                'Tidak ada data validasi (mode v6/degraded).'}
            </pre>
            {approval.validationJson.includes('degraded') && (
              <p className='mt-2 text-xs text-amber-600'>
                Sintaks belum diverifikasi router (v6 degraded). Periksa manual
                sebelum approve.
              </p>
            )}
          </TabsContent>
          <TabsContent value='restore' className='pt-2'>
            <div className='rounded-md border p-3 text-sm'>
              <p>
                Restore point:{' '}
                <span className='font-mono'>
                  {approval.restorePointId || '-'}
                </span>
              </p>
              <p className='mt-1 text-xs text-muted-foreground'>
                Approval hanya bisa dieksekusi bila restore point berstatus
                AVAILABLE.
              </p>
            </div>
          </TabsContent>
        </Tabs>

        {approval.executionResultJson && (
          <div>
            <p className='mb-1 text-xs font-medium text-muted-foreground'>
              Hasil eksekusi
            </p>
            <pre className='max-h-40 overflow-auto rounded-md border bg-muted/30 p-3 font-mono text-xs'>
              {approval.executionResultJson}
            </pre>
          </div>
        )}

        <DialogFooter className='flex-wrap gap-2'>
          {pending && canDecide && (
            <>
              {editMode ? (
                <>
                  <Button variant='outline' onClick={() => setEditMode(false)}>
                    Batal Edit
                  </Button>
                  <Button
                    variant='outline'
                    disabled={decide.isPending}
                    onClick={() => decideWith(ApprovalDecision.EDIT)}
                  >
                    Simpan Edit (PENDING)
                  </Button>
                  <Button
                    disabled={decide.isPending}
                    onClick={() =>
                      decideWith(ApprovalDecision.EDIT_AND_APPROVE)
                    }
                  >
                    Edit & Approve
                  </Button>
                </>
              ) : (
                <>
                  <Button
                    variant='ghost'
                    onClick={() => setEditMode(true)}
                    className='gap-1.5'
                  >
                    <Pencil className='h-4 w-4' /> Edit
                  </Button>
                  <Button
                    variant='outline'
                    onClick={() => setRejectOpen(true)}
                    className='gap-1.5'
                  >
                    <X className='h-4 w-4' /> Reject
                  </Button>
                  <Button
                    onClick={() => decideWith(ApprovalDecision.APPROVE)}
                    disabled={decide.isPending}
                    className='gap-1.5'
                  >
                    {decide.isPending ? (
                      <Loader2 className='h-4 w-4 animate-spin' />
                    ) : (
                      <Check className='h-4 w-4' />
                    )}
                    Approve & Execute
                  </Button>
                </>
              )}
            </>
          )}
          {pending && !canDecide && (
            <p className='text-xs text-muted-foreground'>
              Anda tidak memiliki permission approval:decide untuk memutuskan
              request ini.
            </p>
          )}
          {!pending && (
            <Button variant='outline' onClick={() => onOpenChange(false)}>
              Tutup
            </Button>
          )}
        </DialogFooter>

        <Dialog open={rejectOpen} onOpenChange={setRejectOpen}>
          <DialogContent className='sm:max-w-md'>
            <DialogHeader>
              <DialogTitle>Tolak Approval</DialogTitle>
              <DialogDescription>
                Alasan wajib diisi dan akan dicatat pada audit trail.
              </DialogDescription>
            </DialogHeader>
            <Textarea
              rows={3}
              value={reason}
              onChange={(e) => setReason(e.target.value)}
              placeholder='Alasan penolakan…'
            />
            <DialogFooter>
              <Button variant='outline' onClick={() => setRejectOpen(false)}>
                Batal
              </Button>
              <Button
                variant='destructive'
                disabled={!reason.trim() || decide.isPending}
                onClick={async () => {
                  await decideWith(ApprovalDecision.REJECT, { reason })
                  setRejectOpen(false)
                }}
              >
                Tolak
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </DialogContent>
    </Dialog>
  )
}
