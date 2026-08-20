import React from 'react'
import { FileText, FileCode, Trash2 } from 'lucide-react'
import { cn } from '@/lib/utils'
import { SkillFile } from '../types'

interface SkillFileTreeItemProps {
  file: SkillFile
  isActive: boolean
  onSelect: () => void
  onDelete?: () => void
}

export const SkillFileTreeItem: React.FC<SkillFileTreeItemProps> = ({
  file,
  isActive,
  onSelect,
  onDelete,
}) => {
  const isSpecial = file.name === 'SKILL.md' || file.name === 'system-prompt.md'

  return (
    <div
      onClick={onSelect}
      className={cn(
        'group flex w-full cursor-pointer items-center justify-between rounded-md px-2.5 py-1.5 text-xs transition-colors',
        isActive
          ? 'bg-primary/10 font-medium text-primary'
          : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground'
      )}
    >
      <div className='flex min-w-0 items-center gap-2'>
        {isSpecial ? (
          <FileCode
            size={14}
            className={cn(
              'shrink-0',
              isActive ? 'text-primary' : 'text-amber-500/80'
            )}
          />
        ) : (
          <FileText size={14} className='shrink-0 text-muted-foreground' />
        )}
        <span className='truncate font-mono'>{file.name}</span>
      </div>

      <div className='flex items-center gap-1 opacity-0 transition-opacity group-hover:opacity-100'>
        {onDelete && (
          <button
            type='button'
            onClick={(e) => {
              e.stopPropagation()
              onDelete()
            }}
            className='rounded p-0.5 text-muted-foreground hover:bg-destructive/10 hover:text-destructive'
            title={file.name === 'SKILL.md' ? 'Hapus seluruh skill ini' : 'Hapus berkas ini'}
          >
            <Trash2 size={12} />
          </button>
        )}
      </div>
    </div>
  )
}
