import { ContentSection } from '../components/content-section'
import { AccountForm } from './account-form'

export function SettingsAccount() {
  return (
    <ContentSection
      title='Keamanan & Password Akun'
      desc='Kelola keamanan akun Anda dan perbarui kata sandi secara berkala.'
    >
      <AccountForm />
    </ContentSection>
  )
}
