import { useI18n } from '@/context/i18n-provider'
import { ContentSection } from '../components/content-section'
import { DisplayForm } from './display-form'

export function SettingsDisplay() {
  const { t } = useI18n()
  return (
    <ContentSection
      title={t('settings.display_title')}
      desc={t('settings.display_desc')}
    >
      <DisplayForm />
    </ContentSection>
  )
}
