'use client'

import { DeleteKnowledgeRequest } from '@/gen/v1/knowledge_pb'
import { toast } from 'sonner'
import {
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { Button } from '@/components/ui/button'
import { useDeleteKnowledgeMutation } from '../api/use-knowledge'
import { useKnowledgeContext } from './knowledge-provider'

export function KnowledgeDeleteDialog() {
  const { open, setOpen, currentRow, setCurrentRow } = useKnowledgeContext()
  const deleteMutation = useDeleteKnowledgeMutation()

  async function onDelete() {
    if (!currentRow) return
    try {
      const res = await deleteMutation.mutateAsync(
        new DeleteKnowledgeRequest({ id: currentRow.id })
      )
      toast.success(res.message || 'Document deleted successfully')
      setOpen(null)
      setCurrentRow(null)
    } catch (err: unknown) {
      const errorMessage =
        err instanceof Error ? err.message : 'Failed to delete document'
      toast.error(errorMessage)
    }
  }

  return (
    <AlertDialog
      open={open === 'delete'}
      onOpenChange={(v) => !v && setOpen(null)}
    >
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Are you absolutely sure?</AlertDialogTitle>
          <AlertDialogDescription>
            This action cannot be undone. The document{' '}
            <strong>{currentRow?.title}</strong> will be removed from the
            knowledge base
            {currentRow?.embedToLlm
              ? ' and from the AnythingLLM vector store'
              : ''}
            .
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>Cancel</AlertDialogCancel>
          <Button
            variant='destructive'
            onClick={onDelete}
            disabled={deleteMutation.isPending}
          >
            {deleteMutation.isPending ? 'Deleting...' : 'Delete'}
          </Button>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
