import { useState } from 'react'
import { Loader2, type LucideIcon } from 'lucide-react'
import { toast } from 'sonner'
import { useI18n } from '@/context/i18n-provider'
import { Button } from '@/components/ui/button'

interface OltJobButtonProps {
  label: string
  icon: LucideIcon
  /** Runs a job-based RPC and resolves with the queued job id. */
  onRun: () => Promise<string>
  variant?: 'default' | 'outline' | 'secondary' | 'ghost'
  size?: 'default' | 'sm'
  className?: string
  successMessage?: string
  errorMessage?: string
}

/**
 * Tombol aksi async job queue OLT: menampilkan toast berisi job_id dan
 * mencegah klik ganda selama request berjalan.
 */
export function OltJobButton({
  label,
  icon: Icon,
  onRun,
  variant = 'outline',
  size = 'default',
  className,
  successMessage,
  errorMessage,
}: OltJobButtonProps) {
  const { t } = useI18n()
  const [isRunning, setIsRunning] = useState(false)

  const handleClick = async () => {
    setIsRunning(true)
    try {
      const jobId = await onRun()
      toast.success(
        successMessage ?? t('olt.job_queued', { jobId: jobId || '-' })
      )
    } catch (err) {
      toast.error(
        errorMessage ??
          (err instanceof Error ? err.message : t('olt.job_failed'))
      )
    } finally {
      setIsRunning(false)
    }
  }

  return (
    <Button
      variant={variant}
      size={size}
      className={className}
      onClick={handleClick}
      disabled={isRunning}
    >
      {isRunning ? (
        <Loader2 className='h-4 w-4 animate-spin' />
      ) : (
        <Icon className='h-4 w-4' />
      )}
      {label}
    </Button>
  )
}
