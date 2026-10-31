import { ContentSection } from '../components/content-section'
import { BillingSettingsForm } from './billing-settings-form'

export function SettingsBilling() {
  return (
    <ContentSection
      title='Billing & Keuangan'
      desc='Atur mode siklus tagihan, tutup buku, PPN default, pengingat, dan operasional gateway. Kebijakan: tanpa denda keterlambatan dan tanpa biaya tambahan pada tagihan.'
    >
      <BillingSettingsForm />
    </ContentSection>
  )
}
