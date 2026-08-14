'use client'

import { useNavigate } from '@tanstack/react-router'
import { type Row } from '@tanstack/react-table'
import { type KnowledgeItem, RetryEmbedRequest } from '@/gen/v1/knowledge_pb'
import { MoreHorizontal, Edit, Trash2, RefreshCw } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { useRetryEmbedMutation } from '../api/use-knowledge'
import { useKnowledgeContext } from './knowledge-provider'

interface KnowledgeRowActionsProps {
  row: Row<KnowledgeItem>
}

export function KnowledgeRowActions({ row }: KnowledgeRowActionsProps) {
  const { setOpen, setCurrentRow } = useKnowledgeContext()
  const navigate = useNavigate()
  const retryEmbedMutation = useRetryEmbedMutation()
  const item = row.original

  async function handleRetryEmbed() {
    try {
      const res = await retryEmbedMutation.mutateAsync(
        new RetryEmbedRequest({ id: item.id })
      )
      if (res.item?.embedStatus === 'embedded') {
        toast.success('Document embedded successfully')
      } else if (res.item?.embedStatus === 'failed') {
        toast.error('Embed failed — check AnythingLLM configuration')
      } else {
        toast.success('Embed sync completed')
      }
    } catch (err: unknown) {
      const errorMessage =
        err instanceof Error ? err.message : 'Failed to retry embed'
      toast.error(errorMessage)
    }
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          variant='ghost'
          className='flex h-8 w-8 p-0 data-[state=open]:bg-muted'
        >
          <MoreHorizontal className='h-4 w-4' />
          <span className='sr-only'>Open menu</span>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align='end' className='w-[160px]'>
        <DropdownMenuItem
          onClick={() =>
            navigate({
              to: '/knowledge/$id/edit',
              params: { id: item.id },
            })
          }
        >
          <Edit className='me-2 h-3.5 w-3.5' />
          Edit
        </DropdownMenuItem>
        <DropdownMenuItem
          onClick={handleRetryEmbed}
          disabled={retryEmbedMutation.isPending}
        >
          <RefreshCw className='me-2 h-3.5 w-3.5 text-blue-500' />
          {retryEmbedMutation.isPending ? 'Syncing...' : 'Retry Embed'}
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem
          className='text-red-600 focus:text-red-600'
          onClick={() => {
            setCurrentRow(item)
            setOpen('delete')
          }}
        >
          <Trash2 className='me-2 h-3.5 w-3.5' />
          Delete
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
