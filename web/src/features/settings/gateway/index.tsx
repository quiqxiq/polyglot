import { ContentSection } from '../components/content-section'
import { GatewaySettingsForm } from './gateway-settings-form'

export function SettingsGateway() {
  return (
    <ContentSection
      title='Payment Gateway'
      desc='Kelola gateway pembayaran aktif, kredensial, dan kanal pembayaran. Kredensial sensitif otomatis dienkripsi saat disimpan.'
    >
      <GatewaySettingsForm />
    </ContentSection>
  )
}
