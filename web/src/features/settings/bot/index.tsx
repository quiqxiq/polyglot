import { ContentSection } from '../components/content-section'
import { BotSettingsForm } from './bot-settings-form'

export function SettingsBot() {
  return (
    <ContentSection
      title='Konfigurasi Operasional Bot & Anti-Spam'
      desc='Kelola parameter anti-spam burst, kuota chat harian, konteks memori LLM, dan pengecualian whitelist secara real-time.'
    >
      <BotSettingsForm />
    </ContentSection>
  )
}
