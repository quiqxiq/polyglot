import { format } from 'date-fns'
import { type ColumnDef } from '@tanstack/react-table'
import { type KnowledgeItem } from '@/gen/v1/knowledge_pb'
import { Badge } from '@/components/ui/badge'
import { DataTableColumnHeader } from '@/components/data-table'
import { KnowledgeRowActions } from './knowledge-row-actions'
import { KnowledgeEmbedStatusBadge } from './knowledge-status-badge'

export const knowledgeColumns: ColumnDef<KnowledgeItem>[] = [
  {
    accessorKey: 'title',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Title' />
    ),
    cell: ({ row }) => (
      <div
        className='max-w-72 truncate font-medium'
        title={row.getValue('title')}
      >
        {row.getValue('title')}
      </div>
    ),
  },
  {
    accessorKey: 'category',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Category' />
    ),
    cell: ({ row }) => {
      const category = row.getValue('category') as string
      if (!category)
        return <span className='text-xs text-muted-foreground'>—</span>
      return (
        <Badge variant='outline' className='capitalize'>
          {category}
        </Badge>
      )
    },
  },
  {
    accessorKey: 'tags',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Tags' />
    ),
    cell: ({ row }) => {
      const tags = (row.getValue('tags') as string[]) || []
      if (tags.length === 0)
        return <span className='text-xs text-muted-foreground'>—</span>
      // Tampilkan maksimal 3 tag di tabel; sisanya dirangkum jadi "+N"
      // (tooltip berisi daftar tag lengkap).
      const visible = tags.slice(0, 3)
      const extra = tags.length - visible.length
      return (
        <div
          className='flex max-w-64 flex-wrap items-center gap-1'
          title={tags.join(', ')}
        >
          {visible.map((tag, i) => (
            <Badge
              key={i}
              variant='outline'
              className='px-1.5 py-0 text-[10px]'
            >
              {tag}
            </Badge>
          ))}
          {extra > 0 && (
            <Badge
              variant='secondary'
              className='px-1.5 py-0 text-[10px]'
              title={tags.join(', ')}
            >
              +{extra}
            </Badge>
          )}
        </div>
      )
    },
  },
  {
    accessorKey: 'embedStatus',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Embed' />
    ),
    cell: ({ row }) => (
      <KnowledgeEmbedStatusBadge status={row.getValue('embedStatus')} />
    ),
  },
  {
    accessorKey: 'anythingllmDocName',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='LLM Doc' />
    ),
    cell: ({ row }) => {
      const docName = (row.getValue('anythingllmDocName') as string) || ''
      const embedded = row.getValue('embedStatus') === 'embedded'
      if (!docName || !embedded)
        return <span className='text-xs text-muted-foreground'>—</span>
      // Tampilkan basename (raw-xxx.json) — path folder sudah jelas dari
      // konteks; nama lengkap ada di tooltip.
      const base = docName.split('/').pop() ?? docName
      return (
        <span
          className='inline-block max-w-52 truncate font-mono text-[11px] text-muted-foreground'
          title={docName}
        >
          {base}
        </span>
      )
    },
  },
  {
    accessorKey: 'updatedAt',
    header: ({ column }) => (
      <DataTableColumnHeader column={column} title='Updated' />
    ),
    cell: ({ row }) => {
      const updatedAt = row.getValue('updatedAt') as string
      if (!updatedAt)
        return <span className='text-xs text-muted-foreground'>—</span>
      return (
        <span className='text-xs text-muted-foreground'>
          {format(new Date(updatedAt), 'dd MMM yyyy, HH:mm')}
        </span>
      )
    },
  },
  {
    id: 'actions',
    cell: ({ row }) => <KnowledgeRowActions row={row} />,
  },
]
