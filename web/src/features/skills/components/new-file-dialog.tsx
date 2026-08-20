import React, { useState } from 'react'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
import { Skill } from '../types'

interface NewFileDialogProps {
  open: boolean
  skillName?: string
  targetSkill?: Skill | null
  onOpenChange: (open: boolean) => void
  onCreateFile: (fileName: string, isReference: boolean) => void
}

export const NewFileDialog: React.FC<NewFileDialogProps> = ({
  open,
  skillName,
  targetSkill,
  onOpenChange,
  onCreateFile,
}) => {
  const [fileName, setFileName] = useState('')
  const [fileType, setFileType] = useState<'root' | 'reference'>('reference')

  const displayName = targetSkill ? targetSkill.name : skillName || 'Skill'

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    let clean = fileName.trim()
    if (!clean) return
    if (!clean.endsWith('.md')) {
      clean += '.md'
    }
    onCreateFile(clean, fileType === 'reference')
    setFileName('')
    setFileType('reference')
    onOpenChange(false)
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-[425px]'>
        <form onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle>Tambah Berkas Baru</DialogTitle>
            <DialogDescription>
              Tambahkan berkas Markdown (.md) baru ke dalam modul{' '}
              <strong className='text-foreground'>{displayName}</strong>.
            </DialogDescription>
          </DialogHeader>

          <div className='grid gap-4 py-4'>
            <div className='grid gap-2'>
              <Label htmlFor='file-name' className='text-xs font-medium'>
                Nama Berkas (.md)
              </Label>
              <Input
                id='file-name'
                placeholder='e.g. faq-tambahan.md'
                value={fileName}
                onChange={(e) => setFileName(e.target.value)}
                autoFocus
                className='text-sm'
              />
              <p className='text-[11px] text-muted-foreground'>
                Format nama berkas menggunakan huruf kecil dan strip (kebab-case).
              </p>
            </div>

            <div className='grid gap-2'>
              <Label className='text-xs font-medium'>Lokasi Penyimpanan</Label>
              <RadioGroup
                value={fileType}
                onValueChange={(val: 'root' | 'reference') => setFileType(val)}
                className='grid gap-2 text-xs'
              >
                <div className='flex items-center space-x-2 rounded-md border p-2.5'>
                  <RadioGroupItem value='reference' id='r-ref' />
                  <Label htmlFor='r-ref' className='cursor-pointer font-normal'>
                    Folder Referensi (<code className='text-[11px] font-mono'>references/</code>)
                    <span className='block text-[10px] text-muted-foreground'>
                      Direkomendasikan untuk dokumen panduan & SOP detail.
                    </span>
                  </Label>
                </div>
                <div className='flex items-center space-x-2 rounded-md border p-2.5'>
                  <RadioGroupItem value='root' id='r-root' />
                  <Label htmlFor='r-root' className='cursor-pointer font-normal'>
                    Root Folder Skill (<code className='text-[11px] font-mono'>/</code>)
                    <span className='block text-[10px] text-muted-foreground'>
                      Untuk berkas di samping SKILL.md.
                    </span>
                  </Label>
                </div>
              </RadioGroup>
            </div>
          </div>

          <DialogFooter>
            <Button
              type='button'
              variant='outline'
              onClick={() => onOpenChange(false)}
            >
              Batal
            </Button>
            <Button type='submit' disabled={!fileName.trim()}>
              Buat Berkas
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
