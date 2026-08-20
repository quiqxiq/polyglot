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
import { Textarea } from '@/components/ui/textarea'

interface NewSkillDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onCreateSkill: (slug: string, name: string, description: string) => void
}

export const NewSkillDialog: React.FC<NewSkillDialogProps> = ({
  open,
  onOpenChange,
  onCreateSkill,
}) => {
  const [name, setName] = useState('')
  const [slug, setSlug] = useState('')
  const [description, setDescription] = useState('')

  const handleNameChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const val = e.target.value
    setName(val)
    if (!slug || slug === name.toLowerCase().replace(/[^a-z0-9]+/g, '-')) {
      setSlug(
        val
          .toLowerCase()
          .trim()
          .replace(/[^a-z0-9]+/g, '-')
          .replace(/^-|-$/g, '')
      )
    }
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!name.trim() || !slug.trim()) return
    onCreateSkill(slug.trim(), name.trim(), description.trim())
    setName('')
    setSlug('')
    setDescription('')
    onOpenChange(false)
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-[480px]'>
        <form onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle>Tambah Skill Baru</DialogTitle>
            <DialogDescription>
              Buat paket instruksi dan referensi SOP baru untuk asisten AI bot.
            </DialogDescription>
          </DialogHeader>

          <div className='grid gap-4 py-4 text-sm'>
            <div className='grid gap-2'>
              <Label htmlFor='skill-name'>Nama Skill</Label>
              <Input
                id='skill-name'
                placeholder='contoh: Layanan Hotspot Voucher'
                value={name}
                onChange={handleNameChange}
                required
              />
            </div>

            <div className='grid gap-2'>
              <Label htmlFor='skill-slug'>Identifier / Slug Folder</Label>
              <Input
                id='skill-slug'
                placeholder='contoh: hotspot-voucher-cs'
                value={slug}
                onChange={(e) => setSlug(e.target.value)}
                required
              />
              <span className='text-xs text-muted-foreground'>
                Akan menjadi nama folder di <code>data/skills/{slug || 'nama-skill'}</code>
              </span>
            </div>

            <div className='grid gap-2'>
              <Label htmlFor='skill-desc'>Deskripsi Pemicu (Trigger Description)</Label>
              <Textarea
                id='skill-desc'
                placeholder='Jelaskan kapan skill ini harus digunakan oleh bot AI...'
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                rows={3}
              />
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
            <Button type='submit' disabled={!name.trim() || !slug.trim()}>
              Buat Skill
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
