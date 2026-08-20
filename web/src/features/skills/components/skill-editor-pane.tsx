import React, { useState, useEffect } from 'react'
import { Eye, Code, Save, ChevronDown, Check, FileText, Bot, Sparkles, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { cn } from '@/lib/utils'
import { Skill, SkillFile, ViewMode } from '../types'
import { MarkdownViewer } from './markdown-viewer'
import { MarkdownCodeEditor } from './markdown-code-editor'

interface SkillEditorPaneProps {
  skill?: Skill | null
  activeSkill?: Skill | null
  activeFile: SkillFile | null
  onSelectFile?: (file: SkillFile) => void
  onSaveFileContent: (fileId: string, newContent: string) => void
  onDeleteFile?: (file: SkillFile) => void
  onDeleteSkill?: (skill: Skill) => void
}

export const SkillEditorPane: React.FC<SkillEditorPaneProps> = ({
  skill,
  activeSkill,
  activeFile,
  onSelectFile,
  onSaveFileContent,
  onDeleteFile,
  onDeleteSkill,
}) => {
  const currentSkill = activeSkill || skill || null
  const [viewMode, setViewMode] = useState<ViewMode>('view')
  const [currentContent, setCurrentContent] = useState(activeFile?.content || '')
  const isDirty = activeFile ? currentContent !== activeFile.content : false

  // Sync content when activeFile changes
  useEffect(() => {
    if (activeFile) {
      setCurrentContent(activeFile.content)
    }
  }, [activeFile?.id, activeFile?.content])

  if (!activeFile) {
    return (
      <div className='flex h-full flex-col items-center justify-center p-8 text-center text-muted-foreground'>
        <div className='flex h-12 w-12 items-center justify-center rounded-full bg-muted/60 mb-3'>
          <Sparkles className='h-6 w-6 text-muted-foreground' />
        </div>
        <h3 className='text-sm font-semibold text-foreground'>Tidak Ada Berkas Terpilih</h3>
        <p className='text-xs text-muted-foreground mt-1 max-w-sm'>
          Pilih berkas dari daftar di sebelah kiri untuk melihat atau mulai mengedit SOP dan panduan bot AI.
        </p>
      </div>
    )
  }

  const handleSave = () => {
    onSaveFileContent(activeFile.id, currentContent)
  }

  return (
    <div className='flex h-full flex-col bg-background'>
      {/* Top Header Bar */}
      <div className='flex h-14 flex-none items-center justify-between border-b px-4 py-2 sm:px-6'>
        {/* Left: File Selector Dropdown or Global File Badge */}
        <div className='flex items-center gap-2.5'>
          {activeFile.isGlobal ? (
            <div className='flex items-center gap-2'>
              <div className='flex items-center gap-1.5 rounded-lg border bg-background px-3 py-1.5 text-sm font-semibold shadow-xs'>
                <Bot size={16} className='text-primary' />
                <span>{activeFile.name}</span>
              </div>
              <Badge variant='secondary' className='text-[11px] font-normal'>
                Global Base Prompt
              </Badge>
            </div>
          ) : currentSkill ? (
            <>
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button
                    variant='outline'
                    size='sm'
                    className='h-9 gap-1.5 px-3 font-semibold text-foreground'
                  >
                    <FileText size={15} className='text-primary' />
                    <span>{activeFile.name}</span>
                    <ChevronDown size={14} className='opacity-50' />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align='start' className='w-64'>
                  {currentSkill.files.map((file) => (
                    <DropdownMenuItem
                      key={file.id}
                      onClick={() => onSelectFile?.(file)}
                      className={cn(
                        'flex items-center justify-between py-2 text-xs',
                        file.id === activeFile.id && 'font-medium bg-accent'
                      )}
                    >
                      <div className='flex items-center gap-2 truncate'>
                        <FileText
                          size={13}
                          className={file.isReference ? 'text-muted-foreground' : 'text-primary'}
                        />
                        <span className='truncate'>{file.name}</span>
                      </div>
                      {file.id === activeFile.id && (
                        <Check size={14} className='text-primary shrink-0' />
                      )}
                    </DropdownMenuItem>
                  ))}
                </DropdownMenuContent>
              </DropdownMenu>

              <span className='hidden text-xs text-muted-foreground sm:inline'>
                di dalam <strong className='font-medium text-foreground'>{currentSkill.name}</strong>
              </span>
            </>
          ) : null}
        </div>

        {/* Right: Switch Mode & Save Button */}
        <div className='flex items-center gap-2 sm:gap-3'>
          {/* Segmented Switch: View Mode vs Edit Mode */}
          <div className='flex items-center rounded-lg border bg-muted/40 p-0.5 text-xs shadow-xs'>
            <button
              type='button'
              onClick={() => setViewMode('view')}
              className={cn(
                'flex items-center gap-1.5 rounded-md px-2.5 py-1 font-medium transition-all',
                viewMode === 'view'
                  ? 'bg-background text-foreground shadow-xs'
                  : 'text-muted-foreground hover:text-foreground'
              )}
            >
              <Eye size={14} />
              <span className='hidden sm:inline'>Pratinjau</span>
            </button>
            <button
              type='button'
              onClick={() => setViewMode('edit')}
              className={cn(
                'flex items-center gap-1.5 rounded-md px-2.5 py-1 font-medium transition-all',
                viewMode === 'edit'
                  ? 'bg-background text-foreground shadow-xs'
                  : 'text-muted-foreground hover:text-foreground'
              )}
            >
              <Code size={14} />
              <span className='hidden sm:inline'>Edit Kode</span>
            </button>
          </div>

          {/* Save Button */}
          <Button
            size='sm'
            onClick={handleSave}
            disabled={!isDirty}
            className='h-8 gap-1.5 px-3 text-xs'
          >
            <Save size={14} />
            <span>Simpan</span>
            {isDirty && (
              <span className='h-1.5 w-1.5 rounded-full bg-amber-400 animate-pulse' />
            )}
          </Button>

          {/* Delete Action Button */}
          {!activeFile.isGlobal && (
            <>
              {activeFile.isReference || activeFile.name !== 'SKILL.md' ? (
                <Button
                  variant='ghost'
                  size='sm'
                  onClick={() => onDeleteFile?.(activeFile)}
                  className='h-8 gap-1.5 px-2 text-xs text-muted-foreground hover:bg-destructive/10 hover:text-destructive'
                  title='Hapus berkas ini'
                >
                  <Trash2 size={14} />
                  <span className='hidden sm:inline'>Hapus Berkas</span>
                </Button>
              ) : currentSkill ? (
                <Button
                  variant='ghost'
                  size='sm'
                  onClick={() => onDeleteSkill?.(currentSkill)}
                  className='h-8 gap-1.5 px-2 text-xs text-destructive hover:bg-destructive/10'
                  title='Hapus skill ini beserta seluruh berkasnya'
                >
                  <Trash2 size={14} />
                  <span className='hidden sm:inline'>Hapus Skill</span>
                </Button>
              ) : null}
            </>
          )}
        </div>
      </div>

      {/* Main Content Area */}
      <div className='flex-1 overflow-auto'>
        {viewMode === 'view' ? (
          <MarkdownViewer content={currentContent} />
        ) : (
          <MarkdownCodeEditor
            content={currentContent}
            onChange={(val) => setCurrentContent(val)}
          />
        )}
      </div>
    </div>
  )
}
