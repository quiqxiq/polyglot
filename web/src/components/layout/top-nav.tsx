import { Link } from '@tanstack/react-router'
import { ChevronDown } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'

export interface TopNavLink {
  title: string
  href?: string
  search?: Record<string, unknown>
  isActive: boolean
  disabled?: boolean
  onClick?: () => void
  icon?: React.ReactNode
}

export type TopNavProps = React.HTMLAttributes<HTMLElement> & {
  links: TopNavLink[]
}

export function TopNav({ className, links, ...props }: TopNavProps) {
  const activeLink = links.find((l) => l.isActive) || links[0]

  return (
    <>
      <DropdownMenu modal={false}>
        <DropdownMenuTrigger asChild>
          <Button
            size='sm'
            variant='outline'
            className={cn('h-8 gap-1.5 px-2.5 text-xs font-medium lg:hidden', className)}
          >
            {activeLink?.icon && <span className='size-3.5'>{activeLink.icon}</span>}
            <span className='max-w-[110px] truncate'>{activeLink?.title || 'Navigasi'}</span>
            <ChevronDown className='size-3.5 opacity-60' />
            <span className='sr-only'>Buka menu navigasi</span>
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent side='bottom' align='start' className='w-48'>
          {links.map(({ title, href, search, isActive, disabled, onClick, icon }) => {
            const hasRouteLink = Boolean(href && !onClick)

            return (
              <DropdownMenuItem
                key={`${title}-${href ?? ''}`}
                disabled={disabled}
                onClick={onClick}
                className={cn(
                  'flex items-center gap-2 cursor-pointer',
                  isActive && 'bg-primary/10 font-semibold text-primary'
                )}
                asChild={hasRouteLink}
              >
                {hasRouteLink ? (
                  <Link
                    // eslint-disable-next-line @typescript-eslint/no-explicit-any
                    to={href as any}
                    // eslint-disable-next-line @typescript-eslint/no-explicit-any
                    search={search as any}
                    disabled={disabled}
                    className={cn(
                      'flex w-full items-center gap-2 text-xs',
                      !isActive && 'text-muted-foreground'
                    )}
                  >
                    {icon && <span className='size-3.5 shrink-0'>{icon}</span>}
                    <span>{title}</span>
                  </Link>
                ) : (
                  <div className='flex w-full items-center gap-2 text-xs'>
                    {icon && <span className='size-3.5 shrink-0'>{icon}</span>}
                    <span>{title}</span>
                  </div>
                )}
              </DropdownMenuItem>
            )
          })}
        </DropdownMenuContent>
      </DropdownMenu>

      <nav
        className={cn(
          'hidden items-center gap-1 lg:flex',
          className
        )}
        {...props}
      >
        {links.map(({ title, href, search, isActive, disabled, onClick, icon }) => {
          const content = (
            <>
              {icon && <span className='shrink-0'>{icon}</span>}
              <span>{title}</span>
            </>
          )

          const navItemClasses = cn(
            'inline-flex items-center gap-1.5 rounded-md px-3 py-1.5 text-sm font-medium transition-colors select-none',
            isActive
              ? 'bg-primary/10 text-primary font-semibold shadow-2xs'
              : 'text-muted-foreground hover:bg-muted/80 hover:text-foreground',
            disabled && 'pointer-events-none opacity-50'
          )

          if (onClick || !href) {
            return (
              <button
                key={`${title}-${href ?? ''}`}
                type='button'
                onClick={onClick}
                disabled={disabled}
                className={cn(navItemClasses, 'cursor-pointer')}
              >
                {content}
              </button>
            )
          }

          return (
            <Link
              key={`${title}-${href}`}
              // eslint-disable-next-line @typescript-eslint/no-explicit-any
              to={href as any}
              // eslint-disable-next-line @typescript-eslint/no-explicit-any
              search={search as any}
              disabled={disabled}
              className={navItemClasses}
            >
              {content}
            </Link>
          )
        })}
      </nav>
    </>
  )
}