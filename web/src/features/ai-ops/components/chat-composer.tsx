import type { KeyboardEvent } from 'react'
import { SendIcon, SquareIcon } from 'lucide-react'
import {
  InputGroup,
  InputGroupAddon,
  InputGroupButton,
  InputGroupTextarea,
} from '@/components/ui/input-group'

export interface ChatComposerProps {
  value: string
  onValueChange: (value: string) => void
  onSend: () => void
  onStop: () => void
  busy: boolean
  disabled: boolean
}

export function ChatComposer({
  value,
  onValueChange,
  onSend,
  onStop,
  busy,
  disabled,
}: ChatComposerProps) {
  function handleKeyDown(event: KeyboardEvent<HTMLTextAreaElement>) {
    if (event.key === 'Enter' && !event.shiftKey) {
      event.preventDefault()
      onSend()
    }
  }

  return (
    <div className='border-t pt-3'>
      <InputGroup className='bg-background'>
        <InputGroupTextarea
          value={value}
          onChange={(event) => onValueChange(event.target.value)}
          onKeyDown={handleKeyDown}
          placeholder={
            disabled
              ? 'Pilih router dulu di sidebar'
              : 'Tanya AI Ops… (Enter untuk kirim, Shift+Enter baris baru)'
          }
          disabled={disabled || busy}
          className='min-h-[52px] resize-none'
        />
        <InputGroupAddon
          align='block-end'
          className='justify-between border-t px-2 py-1.5'
        >
          <div className='text-[11px] text-muted-foreground'>
            Enter untuk kirim · Shift+Enter untuk baris baru
          </div>
          {busy ? (
            <InputGroupButton
              variant='outline'
              size='sm'
              onClick={onStop}
              className='gap-1.5'
            >
              <SquareIcon className='size-3.5' /> Stop
            </InputGroupButton>
          ) : (
            <InputGroupButton
              variant='default'
              size='sm'
              onClick={onSend}
              disabled={disabled || !value.trim()}
              className='gap-1.5'
            >
              <SendIcon className='size-3.5' /> Kirim
            </InputGroupButton>
          )}
        </InputGroupAddon>
      </InputGroup>
    </div>
  )
}
