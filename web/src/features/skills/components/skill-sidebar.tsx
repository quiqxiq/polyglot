import React, { useState } from 'react'
import {
  Search,
  Plus,
  Folder,
  FolderOpen,
  Sparkles,
  ChevronRight,
  ChevronDown,
  FilePlus,
  Bot,
  Trash2,
} from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
import { Badge } from '@/components/ui/badge'
import { ScrollArea } from '@/components/ui/scroll-area'
import { cn } from '@/lib/utils'
import { Skill, SkillFile } from '../types'
import { SkillFileTreeItem } from './skill-file-tree-item'

interface SkillSidebarProps {
  skills: Skill[]
  activeSkill: Skill | null
  activeFile: SkillFile | null
  globalPrompt?: SkillFile
  globalSystemPrompt?: SkillFile
  onSelectGlobalPrompt: () => void
  onSelectSkill: (skill: Skill) => void
  onSelectFile: (skill: Skill, file: SkillFile) => void
  onToggleSkillEnabled: (skillId: string, enabled: boolean) => void
  onOpenNewSkill?: () => void
  onOpenNewSkillDialog?: () => void
  onOpenNewFile?: (skill: Skill) => void
  onOpenNewFileDialog?: (skill: Skill) => void
  onDeleteFile?: (skillId: string, fileId: string) => void
  onDeleteSkill?: (skill: Skill) => void
}

export const SkillSidebar: React.FC<SkillSidebarProps> = ({
  skills,
  activeSkill,
  activeFile,
  globalPrompt,
  globalSystemPrompt,
  onSelectGlobalPrompt,
  onSelectSkill,
  onSelectFile,
  onToggleSkillEnabled,
  onOpenNewSkill,
  onOpenNewSkillDialog,
  onOpenNewFile,
  onOpenNewFileDialog,
  onDeleteFile,
  onDeleteSkill,
}) => {
  const effectiveGlobalPrompt = globalSystemPrompt || globalPrompt!
  const triggerNewSkill = onOpenNewSkillDialog || onOpenNewSkill || (() => {})
  const triggerNewFile = onOpenNewFileDialog || onOpenNewFile || (() => {})
  const [search, setSearch] = useState('')
  const [expandedSkills, setExpandedSkills] = useState<Record<string, boolean>>({
    'skill-ghaib-network-cs': true,
    'skill-voucher-sales': true,
  })
  const [expandedFolders, setExpandedFolders] = useState<Record<string, boolean>>({
    'skill-ghaib-network-cs-refs': true,
    'skill-voucher-sales-refs': true,
  })

  const toggleSkillExpand = (skillId: string) => {
    setExpandedSkills((prev) => ({
      ...prev,
      [skillId]: !prev[skillId],
    }))
  }

  const toggleFolderExpand = (folderKey: string) => {
    setExpandedFolders((prev) => ({
      ...prev,
      [folderKey]: !prev[folderKey],
    }))
  }

  const filteredSkills = skills.filter(
    (s) =>
      s.name.toLowerCase().includes(search.toLowerCase()) ||
      s.slug.toLowerCase().includes(search.toLowerCase()) ||
      s.files.some((f) => f.name.toLowerCase().includes(search.toLowerCase()))
  )

  return (
    <div className='flex h-full w-full flex-col border-r bg-card'>
      {/* Top Header */}
      <div className='flex flex-none flex-col gap-3 border-b p-4'>
        <div className='flex items-center justify-between'>
          <div className='flex items-center gap-2'>
            <div className='flex h-8 w-8 items-center justify-center rounded-lg bg-primary/10 text-primary'>
              <Sparkles size={18} />
            </div>
            <div>
              <h2 className='text-sm font-semibold'>Skills & Prompts</h2>
              <p className='text-[11px] text-muted-foreground'>
                {skills.length} skills terpasang
              </p>
            </div>
          </div>

          <Button
            size='sm'
            onClick={triggerNewSkill}
            className='h-8 gap-1.5 px-2.5 text-xs'
          >
            <Plus size={14} />
            <span>Skill</span>
          </Button>
        </div>

        {/* Search Input */}
        <div className='relative'>
          <Search
            size={14}
            className='absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground'
          />
          <Input
            placeholder='Cari skill atau berkas...'
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className='h-8 pl-8 text-xs'
          />
        </div>
      </div>

      {/* Main List Area */}
      <ScrollArea className='flex-1 p-3'>
        {/* 1. Global Base System Prompt (DI LUAR FOLDER SKILL) */}
        <div className='mb-4'>
          <div className='mb-1.5 px-1 text-[10px] font-semibold tracking-wider text-muted-foreground uppercase'>
            System Prompt (Global)
          </div>
          <div
            onClick={onSelectGlobalPrompt}
            className={cn(
              'group flex cursor-pointer items-center justify-between rounded-lg border p-2.5 transition-all',
              activeFile?.isGlobal
                ? 'border-primary/50 bg-primary/10 font-medium text-primary shadow-xs'
                : 'border-border/60 bg-card hover:border-border hover:bg-accent/40 text-foreground'
            )}
          >
            <div className='flex min-w-0 items-center gap-2.5'>
              <div
                className={cn(
                  'flex h-7 w-7 shrink-0 items-center justify-center rounded-md transition-colors',
                  activeFile?.isGlobal
                    ? 'bg-primary text-primary-foreground'
                    : 'bg-primary/10 text-primary group-hover:bg-primary/15'
                )}
              >
                <Bot size={15} />
              </div>
              <div className='min-w-0 flex-1'>
                <div className='truncate text-xs font-semibold'>
                  {effectiveGlobalPrompt?.name || 'system-prompt.md'}
                </div>
                <div className='truncate text-[10px] font-mono text-muted-foreground'>
                  {effectiveGlobalPrompt?.path || effectiveGlobalPrompt?.filePath || 'system-prompt.md'}
                </div>
              </div>
            </div>
            <Badge
              variant='outline'
              className='shrink-0 text-[10px] font-normal uppercase text-muted-foreground'
            >
              Root
            </Badge>
          </div>
        </div>

        {/* 2. Modular Skills Section (DI DALAM FOLDER SKILLS) */}
        <div>
          <div className='mb-1.5 flex items-center justify-between px-1 text-[10px] font-semibold tracking-wider text-muted-foreground uppercase'>
            <span>Modular Skills (data/skills/)</span>
            <span>{filteredSkills.length}</span>
          </div>

          {filteredSkills.length === 0 ? (
            <p className='p-4 text-center text-xs text-muted-foreground'>
              Tidak ada skill yang cocok dengan pencarian.
            </p>
          ) : (
            <div className='space-y-3'>
              {filteredSkills.map((skill) => {
                const isExpanded = expandedSkills[skill.id] ?? true
                const isCurrentSkillActive =
                  !activeFile?.isGlobal && activeSkill?.id === skill.id
                const rootFiles = skill.files.filter((f) => !f.isReference)
                const refFiles = skill.files.filter((f) => f.isReference)
                const folderKey = `${skill.id}-refs`
                const isRefFolderExpanded = expandedFolders[folderKey] ?? true

                return (
                  <div
                    key={skill.id}
                    className={cn(
                      'rounded-lg border p-2 transition-colors',
                      isCurrentSkillActive
                        ? 'border-primary/40 bg-accent/30'
                        : 'border-border/60 bg-card hover:border-border'
                    )}
                  >
                    {/* Skill Header */}
                    <div className='flex items-center justify-between gap-1'>
                      <button
                        type='button'
                        onClick={() => {
                          onSelectSkill(skill)
                          toggleSkillExpand(skill.id)
                        }}
                        className='flex min-w-0 flex-1 items-center gap-1.5 text-left'
                      >
                        {isExpanded ? (
                          <ChevronDown
                            size={14}
                            className='shrink-0 text-muted-foreground'
                          />
                        ) : (
                          <ChevronRight
                            size={14}
                            className='shrink-0 text-muted-foreground'
                          />
                        )}
                        <div className='min-w-0 flex-1'>
                          <div className='truncate text-xs font-semibold'>
                            {skill.name}
                          </div>
                          <div className='truncate text-[10px] font-mono text-muted-foreground'>
                            {skill.slug}
                          </div>
                        </div>
                      </button>

                      <div className='flex items-center gap-1'>
                        <button
                          type='button'
                          onClick={() => triggerNewFile(skill)}
                          className='rounded p-1 text-muted-foreground transition-colors hover:bg-accent hover:text-foreground'
                          title='Tambah berkas ke skill ini'
                        >
                          <FilePlus size={14} />
                        </button>
                        <button
                          type='button'
                          onClick={(e) => {
                            e.stopPropagation()
                            onDeleteSkill?.(skill)
                          }}
                          className='rounded p-1 text-muted-foreground transition-colors hover:bg-destructive/10 hover:text-destructive'
                          title='Hapus skill ini'
                        >
                          <Trash2 size={13} />
                        </button>
                        <Switch
                          checked={skill.enabled}
                          onCheckedChange={(checked) =>
                            onToggleSkillEnabled(skill.id, checked)
                          }
                          className='scale-75'
                          title={
                            skill.enabled ? 'Skill Aktif' : 'Skill Nonaktif'
                          }
                        />
                      </div>
                    </div>

                    {/* Expandable File Tree */}
                    {isExpanded && (
                      <div className='mt-2 space-y-0.5 border-t pt-1.5 pl-3'>
                        {/* Root files */}
                        {rootFiles.map((file) => (
                          <SkillFileTreeItem
                            key={file.id}
                            file={file}
                            isActive={
                              !activeFile?.isGlobal &&
                              activeFile?.id === file.id
                            }
                            onSelect={() => onSelectFile(skill, file)}
                            onDelete={() =>
                              file.name === 'SKILL.md'
                                ? onDeleteSkill?.(skill)
                                : onDeleteFile?.(skill.id, file.id)
                            }
                          />
                        ))}

                        {/* References Folder */}
                        {refFiles.length > 0 && (
                          <div className='pt-0.5'>
                            <button
                              type='button'
                              onClick={() => toggleFolderExpand(folderKey)}
                              className='flex w-full items-center gap-1.5 rounded-md px-2 py-1 text-left text-xs font-medium text-muted-foreground hover:bg-accent hover:text-foreground'
                            >
                              {isRefFolderExpanded ? (
                                <FolderOpen
                                  size={13}
                                  className='text-amber-500'
                                />
                              ) : (
                                <Folder size={13} className='text-amber-500' />
                              )}
                              <span className='font-mono text-[11px]'>
                                references/
                              </span>
                              <span className='ml-auto text-[10px] text-muted-foreground'>
                                {refFiles.length}
                              </span>
                            </button>

                            {isRefFolderExpanded && (
                              <div className='space-y-0.5 pl-3'>
                                {refFiles.map((file) => (
                                  <SkillFileTreeItem
                                    key={file.id}
                                    file={file}
                                    isActive={
                                      !activeFile?.isGlobal &&
                                      activeFile?.id === file.id
                                    }
                                    onSelect={() => onSelectFile(skill, file)}
                                    onDelete={() =>
                                      onDeleteFile?.(skill.id, file.id)
                                    }
                                  />
                                ))}
                              </div>
                            )}
                          </div>
                        )}
                      </div>
                    )}
                  </div>
                )
              })}
            </div>
          )}
        </div>
      </ScrollArea>
    </div>
  )
}
