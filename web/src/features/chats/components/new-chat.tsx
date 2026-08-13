import { useState } from 'react'
import { Send } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'

type NewChatProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSend: (phone: string, messageText: string) => void
  pending?: boolean
}

export function NewChat({ open, onOpenChange, onSend, pending }: NewChatProps) {
  const [phone, setPhone] = useState('')
  const [messageText, setMessageText] = useState('')

  const handleOpenChange = (newOpen: boolean) => {
    onOpenChange(newOpen)
    if (!newOpen) {
      setPhone('')
      setMessageText('')
    }
  }

  const handleSend = () => {
    const clean = phone.trim()
    if (!clean) return
    onSend(clean, messageText.trim())
    setPhone('')
    setMessageText('')
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className='sm:max-w-120'>
        <DialogHeader>
          <DialogTitle>New message</DialogTitle>
        </DialogHeader>
        <div className='flex flex-col gap-4'>
          <div className='flex flex-col gap-1.5'>
            <label htmlFor='new-chat-phone' className='text-sm text-muted-foreground'>
              Nomor WhatsApp penerima
            </label>
            <Input
              id='new-chat-phone'
              type='tel'
              placeholder='e.g. 6281234567890'
              value={phone}
              onChange={(e) => setPhone(e.target.value)}
            />
          </div>
          <div className='flex flex-col gap-1.5'>
            <label htmlFor='new-chat-message' className='text-sm text-muted-foreground'>
              Pesan awal (opsional)
            </label>
            <Input
              id='new-chat-message'
              placeholder='Halo...'
              value={messageText}
              onChange={(e) => setMessageText(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter') handleSend()
              }}
            />
          </div>
          <Button variant='default' onClick={handleSend} disabled={!phone.trim() || pending}>
            <Send size={16} className='me-1' /> Chat
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  )
}
