import { useState, useEffect } from 'react'
import { Sparkles, ArrowLeft, Cpu, Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { cn } from '@/lib/utils'
import { Skill, SkillFile } from './types'
import { SkillSidebar } from './components/skill-sidebar'
import { SkillEditorPane } from './components/skill-editor-pane'
import { NewSkillDialog } from './components/new-skill-dialog'
import { NewFileDialog } from './components/new-file-dialog'
import { LLMConfigDialog } from './components/llm-config-dialog'
import { ConfirmDialog } from '@/components/confirm-dialog'
import {
  useSkills,
  useGlobalPrompt,
  useSaveGlobalPrompt,
  useSaveSkillFile,
  useCreateSkill,
  useDeleteSkill,
  useDeleteSkillFile,
  useToggleSkill,
} from './api/use-skills'
import { botClient } from '@/lib/api-client'

export function Skills() {
  const { data: serverSkills, isLoading: isSkillsLoading } = useSkills()
  const { data: serverGlobalPrompt, isLoading: isPromptLoading } = useGlobalPrompt()

  const saveGlobalPromptMutation = useSaveGlobalPrompt()
  const saveSkillFileMutation = useSaveSkillFile()
  const createSkillMutation = useCreateSkill()
  const deleteSkillMutation = useDeleteSkill()
  const deleteSkillFileMutation = useDeleteSkillFile()
  const toggleSkillMutation = useToggleSkill()

  const [activeSkill, setActiveSkill] = useState<Skill | null>(null)
  const [activeFile, setActiveFile] = useState<SkillFile | null>(null)
  const [mobileSelected, setMobileSelected] = useState<boolean>(false)

  // Dialogs
  const [isNewSkillOpen, setIsNewSkillOpen] = useState(false)
  const [isNewFileOpen, setIsNewFileOpen] = useState(false)
  const [isLLMConfigOpen, setIsLLMConfigOpen] = useState(false)
  const [targetSkillForNewFile, setTargetSkillForNewFile] = useState<Skill | null>(null)

  // Confirmation Modals
  const [skillToDelete, setSkillToDelete] = useState<Skill | null>(null)
  const [fileToDelete, setFileToDelete] = useState<{ skill: Skill; file: SkillFile } | null>(null)

  // Sinkronisasi file aktif saat data termuat pertama kali
  useEffect(() => {
    if (!activeFile) {
      if (serverGlobalPrompt) {
        setActiveFile({
          id: 'global-system-prompt',
          name: 'system-prompt.md',
          path: 'data/system-prompt.md',
          filePath: 'data/system-prompt.md',
          content: serverGlobalPrompt,
          isGlobal: true,
        })
      } else if (serverSkills && serverSkills.length > 0) {
        setActiveSkill(serverSkills[0])
        if (serverSkills[0].files.length > 0) {
          setActiveFile(serverSkills[0].files[0])
        }
      }
    } else if (activeFile.isGlobal && serverGlobalPrompt !== undefined) {
      setActiveFile((prev) =>
        prev
          ? {
              ...prev,
              content: serverGlobalPrompt,
            }
          : null
      )
    }
  }, [serverSkills, serverGlobalPrompt])

  const globalPromptFile: SkillFile = {
    id: 'global-system-prompt',
    name: 'system-prompt.md',
    path: 'data/system-prompt.md',
    filePath: 'data/system-prompt.md',
    content: serverGlobalPrompt || '',
    isGlobal: true,
  }

  const handleSelectGlobalPrompt = () => {
    setActiveSkill(null)
    setActiveFile(globalPromptFile)
    setMobileSelected(true)
  }

  const handleSelectSkill = (skill: Skill) => {
    setActiveSkill(skill)
    if (!skill.files.some((f) => f.id === activeFile?.id)) {
      setActiveFile(skill.files[0] || null)
    }
  }

  const handleSelectFile = async (skill: Skill, file: SkillFile) => {
    setActiveSkill(skill)
    if (!file.content && file.path !== 'SKILL.md' && !file.isGlobal) {
      try {
        const resp = await botClient.getResource({
          skillId: skill.id,
          path: file.filePath || file.path,
        })
        const loadedFile = { ...file, content: resp.content?.content || '' }
        setActiveFile(loadedFile)
        setMobileSelected(true)
        return
      } catch {
        // Continue with original file object if fetch fails
      }
    }
    setActiveFile(file)
    setMobileSelected(true)
  }

  const handleToggleSkillEnabled = (slug: string, enabled: boolean) => {
    toggleSkillMutation.mutate({ slug, enabled })
  }

  const handleSaveFileContent = async (fileId: string, newContent: string) => {
    if (activeFile?.isGlobal || fileId === 'global-system-prompt') {
      await saveGlobalPromptMutation.mutateAsync(newContent)
      setActiveFile((prev) => (prev ? { ...prev, content: newContent } : null))
      return
    }

    if (activeSkill && activeFile) {
      await saveSkillFileMutation.mutateAsync({
        slug: activeSkill.slug,
        filePath: activeFile.filePath || activeFile.path,
        content: newContent,
        isReference: !!activeFile.isReference,
      })
      setActiveFile((prev) => (prev ? { ...prev, content: newContent } : null))
    }
  }

  const handleCreateSkill = async (
    slug: string,
    name: string,
    description: string
  ) => {
    await createSkillMutation.mutateAsync({ slug, name, description })
  }

  const handleCreateFile = async (
    fileName: string,
    isReference: boolean
  ) => {
    if (!targetSkillForNewFile) return
    const filePath = isReference ? `references/${fileName}` : fileName
    const initialContent = `# ${fileName}\n\nDokumen referensi pendukung untuk ${targetSkillForNewFile.name}.\n`

    await saveSkillFileMutation.mutateAsync({
      slug: targetSkillForNewFile.slug,
      filePath,
      content: initialContent,
      isReference,
    })
  }

  const handleDeleteSkill = (skill: Skill) => {
    setSkillToDelete(skill)
  }

  const handleConfirmDeleteSkill = async () => {
    if (!skillToDelete) return
    try {
      await deleteSkillMutation.mutateAsync({
        id: Number(skillToDelete.id) || 0,
        slug: skillToDelete.slug,
      })
      if (activeSkill?.id === skillToDelete.id) {
        setActiveSkill(null)
        setActiveFile(globalPromptFile)
      }
      setSkillToDelete(null)
    } catch {
      // Error handled by mutation
    }
  }

  const handleDeleteFile = (skillId: string, fileId: string) => {
    const sk = (serverSkills || []).find((s) => s.id === skillId)
    if (!sk) return
    const f = sk.files.find((file) => file.id === fileId)
    if (!f) return
    setFileToDelete({ skill: sk, file: f })
  }

  const handleConfirmDeleteFile = async () => {
    if (!fileToDelete) return
    try {
      await deleteSkillFileMutation.mutateAsync({
        slug: fileToDelete.skill.slug,
        fileId: Number(fileToDelete.file.id) || 0,
        filePath: fileToDelete.file.filePath || fileToDelete.file.path,
      })
      if (activeFile?.id === fileToDelete.file.id) {
        const remaining = fileToDelete.skill.files.filter(
          (item) => item.id !== fileToDelete.file.id
        )
        setActiveFile(remaining[0] || globalPromptFile)
      }
      setFileToDelete(null)
    } catch {
      // Error handled by mutation
    }
  }

  return (
    <>
      <Header>
        <div className='flex items-center gap-2'>
          <Sparkles className='h-5 w-5 text-primary' />
          <h1 className='text-lg font-semibold tracking-tight'>
            Skills & System Prompts
          </h1>
        </div>
        <div className='ml-auto flex items-center space-x-2'>
          <Button
            variant='outline'
            size='sm'
            onClick={() => setIsLLMConfigOpen(true)}
            className='gap-1.5'
          >
            <Cpu className='h-4 w-4 text-primary' />
            <span className='hidden sm:inline'>Pengaturan LLM</span>
          </Button>

          <Search />
          <ThemeSwitch />
          <ProfileDropdown />
        </div>
      </Header>

      <Main className='p-0 overflow-hidden'>
        <div className='flex h-[calc(100vh-4rem)] w-full overflow-hidden bg-background'>
          {/* SISI KIRI: File Tree & Daftar Skill */}
          <div
            className={cn(
              'w-full border-r md:w-80 md:min-w-[20rem] md:max-w-sm shrink-0 md:flex flex-col',
              mobileSelected ? 'hidden md:flex' : 'flex'
            )}
          >
            {isSkillsLoading || isPromptLoading ? (
              <div className='flex h-full items-center justify-center'>
                <Loader2 className='h-6 w-6 animate-spin text-muted-foreground' />
              </div>
            ) : (
              <SkillSidebar
                globalPrompt={globalPromptFile}
                skills={serverSkills || []}
                activeSkill={activeSkill}
                activeFile={activeFile}
                onSelectGlobalPrompt={handleSelectGlobalPrompt}
                onSelectSkill={handleSelectSkill}
                onSelectFile={handleSelectFile}
                onToggleSkillEnabled={(id, enabled) => {
                  const s = (serverSkills || []).find((item) => item.id === id)
                  if (s) handleToggleSkillEnabled(s.slug, enabled)
                }}
                onOpenNewSkill={() => setIsNewSkillOpen(true)}
                onOpenNewFile={(skill: Skill) => {
                  setTargetSkillForNewFile(skill)
                  setIsNewFileOpen(true)
                }}
                onDeleteFile={handleDeleteFile}
                onDeleteSkill={handleDeleteSkill}
              />
            )}
          </div>

          {/* SISI KANAN: Dual-Mode Editor (View & Code Edit) */}
          <div
            className={cn(
              'flex-1 flex-col overflow-hidden',
              mobileSelected ? 'flex' : 'hidden md:flex'
            )}
          >
            {mobileSelected && (
              <div className='border-b p-2 md:hidden bg-muted/40'>
                <Button
                  variant='ghost'
                  size='sm'
                  onClick={() => setMobileSelected(false)}
                  className='gap-1.5 text-xs'
                >
                  <ArrowLeft className='h-4 w-4' /> Kembali ke Daftar File
                </Button>
              </div>
            )}

            <SkillEditorPane
              activeFile={activeFile}
              activeSkill={activeSkill}
              onSelectFile={(f) => setActiveFile(f)}
              onSaveFileContent={handleSaveFileContent}
              onDeleteFile={(file) => {
                if (activeSkill) {
                  handleDeleteFile(activeSkill.id, file.id)
                }
              }}
              onDeleteSkill={handleDeleteSkill}
            />
          </div>
        </div>
      </Main>

      {/* Modal Dialogs */}
      <NewSkillDialog
        open={isNewSkillOpen}
        onOpenChange={setIsNewSkillOpen}
        onCreateSkill={handleCreateSkill}
      />

      <NewFileDialog
        open={isNewFileOpen}
        onOpenChange={setIsNewFileOpen}
        targetSkill={targetSkillForNewFile}
        onCreateFile={handleCreateFile}
      />

      <LLMConfigDialog
        open={isLLMConfigOpen}
        onOpenChange={setIsLLMConfigOpen}
      />

      {/* CONFIRM DIALOG: HAPUS SKILL */}
      <ConfirmDialog
        open={!!skillToDelete}
        onOpenChange={(open) => {
          if (!open) setSkillToDelete(null)
        }}
        title='Hapus Modul Skill'
        desc={
          <div className='space-y-2 text-xs sm:text-sm text-muted-foreground'>
            <p>
              Apakah Anda yakin ingin menghapus modul skill{' '}
              <strong className='font-semibold text-foreground'>
                {skillToDelete?.name}
              </strong>{' '}
              (<code className='font-mono text-xs'>{skillToDelete?.slug}</code>)?
            </p>
            <div className='rounded-md bg-destructive/10 p-2.5 text-xs text-destructive'>
              <strong>Peringatan:</strong> Seluruh berkas SOP di folder{' '}
              <code className='font-mono font-semibold'>data/skills/{skillToDelete?.slug}</code> dan
              seluruh dokumen referensi di dalamnya akan dihapus secara permanen dari server dan
              database.
            </div>
          </div>
        }
        confirmText='Hapus Modul Skill'
        cancelBtnText='Batal'
        destructive
        isLoading={deleteSkillMutation.isPending}
        handleConfirm={handleConfirmDeleteSkill}
      />

      {/* CONFIRM DIALOG: HAPUS BERKAS REFERENSI */}
      <ConfirmDialog
        open={!!fileToDelete}
        onOpenChange={(open) => {
          if (!open) setFileToDelete(null)
        }}
        title='Hapus Berkas SOP / Referensi'
        desc={
          <div className='space-y-2 text-xs sm:text-sm text-muted-foreground'>
            <p>
              Apakah Anda yakin ingin menghapus berkas{' '}
              <strong className='font-semibold text-foreground'>
                {fileToDelete?.file.name}
              </strong>{' '}
              dari skill{' '}
              <strong className='font-semibold text-foreground'>
                {fileToDelete?.skill.name}
              </strong>?
            </p>
            <p className='text-xs'>
              Berkas fisik di disk server akan dihapus secara permanen.
            </p>
          </div>
        }
        confirmText='Hapus Berkas'
        cancelBtnText='Batal'
        destructive
        isLoading={deleteSkillFileMutation.isPending}
        handleConfirm={handleConfirmDeleteFile}
      />
    </>
  )
}
