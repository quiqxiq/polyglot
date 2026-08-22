import { ArrowDown } from 'lucide-react'
import { Button } from '@/components/ui/button'

interface JumpToLatestButtonProps {
  visible: boolean
  onClick: () => void
  label?: string
}

export function JumpToLatestButton({
  visible,
  onClick,
  label = 'Jump to Latest',
}: JumpToLatestButtonProps) {
  if (!visible) return null

  return (
    <div className='fixed bottom-6 right-8 z-50 animate-in fade-in slide-in-from-bottom-3 duration-200 pointer-events-auto select-none'>
      <Button
        size='sm'
        onClick={onClick}
        className='h-9 shadow-2xl bg-primary text-primary-foreground hover:bg-primary/90 border border-primary/20 text-xs font-semibold gap-2 px-4 rounded-full transition-all duration-200 hover:scale-105 active:scale-95'
      >
        <ArrowDown className='size-3.5 animate-bounce' />
        <span>{label}</span>
      </Button>
    </div>
  )
}
