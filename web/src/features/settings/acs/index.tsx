import { useI18n } from '@/context/i18n-provider'
import { useCan } from '@/hooks/use-can'
import { AcsConfigCard } from '@/features/ont/components/acs/acs-config-card'
import { AcsDevicesTable } from '@/features/ont/components/acs/acs-devices-table'

/**
 * Section ACS / TR-069 di halaman Settings: status konfigurasi GenieACS,
 * katalog device TR-069, dan pin/unpin device ke ONT.
 */
export function SettingsAcs() {
  const { t } = useI18n()
  const canManage = useCan('ont:manage')

  return (
    <div className='flex flex-1 flex-col gap-6'>
      <div>
        <h3 className='text-lg font-medium'>{t('acs.title')}</h3>
        <p className='text-sm text-muted-foreground'>{t('acs.desc')}</p>
      </div>
      <AcsConfigCard canManage={canManage} />
      <AcsDevicesTable canManage={canManage} />
    </div>
  )
}
