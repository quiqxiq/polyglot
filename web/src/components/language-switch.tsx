import { Check, Globe } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useI18n } from '@/context/i18n-provider'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'

export function LanguageSwitch() {
  const { language, setLanguage } = useI18n()

  return (
    <DropdownMenu modal={false}>
      <DropdownMenuTrigger asChild>
        <Button
          variant='ghost'
          size='sm'
          className='scale-95 h-9 gap-1.5 px-2.5 rounded-full font-medium text-xs'
          title='Change language / Ganti bahasa'
        >
          <Globe className='size-4 text-muted-foreground' />
          <span className='uppercase'>{language}</span>
          <span className='sr-only'>Change language</span>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align='end' className='w-44'>
        <DropdownMenuItem
          onClick={() => setLanguage('en')}
          className='flex items-center justify-between cursor-pointer'
        >
          <div className='flex items-center gap-2'>
            <span className='text-base leading-none'>🇺🇸</span>
            <span>English (EN)</span>
          </div>
          <Check
            size={14}
            className={cn('ms-auto', language !== 'en' && 'hidden')}
          />
        </DropdownMenuItem>
        <DropdownMenuItem
          onClick={() => setLanguage('id')}
          className='flex items-center justify-between cursor-pointer'
        >
          <div className='flex items-center gap-2'>
            <span className='text-base leading-none'>🇮🇩</span>
            <span>Indonesia (ID)</span>
          </div>
          <Check
            size={14}
            className={cn('ms-auto', language !== 'id' && 'hidden')}
          />
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
