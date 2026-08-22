import React from 'react'
import { cn } from '@/lib/utils'

export interface LogoProps extends React.ImgHTMLAttributes<HTMLImageElement> {
  variant?: 'icon' | 'full' | 'bg'
  className?: string
}

export function Logo({
  variant = 'icon',
  className,
  alt = 'Polyglot Logo',
  ...props
}: LogoProps) {
  const imageSrc =
    variant === 'bg'
      ? '/images/logo-bg.png'
      : '/images/logo.png'

  return (
    <img
      src={imageSrc}
      alt={alt}
      className={cn(
        variant === 'full' ? 'h-8 w-auto object-contain' : 'size-7 object-contain',
        'drop-shadow-[0_0_10px_rgba(0,242,254,0.25)] transition-transform hover:scale-105',
        className
      )}
      {...props}
    />
  )
}

export function PolyglotBrand({
  className,
  showTagline = true,
  size = 'md',
}: {
  className?: string
  showTagline?: boolean
  size?: 'sm' | 'md' | 'lg'
}) {
  const sizeClasses = {
    sm: { img: 'h-6', text: 'text-base', sub: 'text-[9px]' },
    md: { img: 'h-8', text: 'text-lg', sub: 'text-[10px]' },
    lg: { img: 'h-10', text: 'text-xl', sub: 'text-xs' },
  }[size]

  return (
    <div className={cn('flex items-center gap-2.5 select-none', className)}>
      <img
        src='/images/logo.png'
        alt='Polyglot'
        className={cn(sizeClasses.img, 'w-auto object-contain drop-shadow-[0_0_12px_rgba(0,242,254,0.3)]')}
      />
      <div className='flex flex-col leading-tight'>
        <span className={cn('font-extrabold tracking-tight bg-gradient-to-r from-cyan-400 via-sky-500 to-fuchsia-500 bg-clip-text text-transparent', sizeClasses.text)}>
          Polyglot
        </span>
        {showTagline && (
          <span className={cn('uppercase tracking-wider font-semibold text-muted-foreground', sizeClasses.sub)}>
            NetOps & ISP Engine
          </span>
        )}
      </div>
    </div>
  )
}
