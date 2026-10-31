import { Link } from '@tanstack/react-router'
import { Bot, ShieldCheck, ArchiveRestore } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { ConfigDrawer } from '@/components/config-drawer'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'

export function AIOpsHome() {
  return (
    <>
      <Header fixed>
        <Search className='me-auto' />
        <ThemeSwitch />
        <ConfigDrawer />
        <ProfileDropdown />
      </Header>
      <Main className='flex flex-1 flex-col gap-4 sm:gap-6'>
        <div>
          <div className='flex items-center gap-2'>
            <Bot className='h-6 w-6 text-primary' />
            <h1 className='text-2xl font-bold tracking-tight'>AI Router Ops</h1>
          </div>
          <p className='mt-0.5 text-sm text-muted-foreground'>
            AI boleh membaca router sebebas mungkin, tetapi setiap perubahan
            wajib disetujui manusia setelah melihat diff dan restore point.
          </p>
        </div>
        <div className='grid gap-4 md:grid-cols-2'>
          <Card>
            <CardHeader>
              <CardTitle className='flex items-center gap-2 text-base'>
                <ShieldCheck className='h-4 w-4 text-primary' /> Approval Inbox
              </CardTitle>
              <CardDescription>
                Review mutation yang dipropos AI: diff, command, validasi, dan
                restore point.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <Button asChild>
                <Link to='/ai-ops/approvals'>Buka Approval Inbox</Link>
              </Button>
            </CardContent>
          </Card>
          <Card>
            <CardHeader>
              <CardTitle className='flex items-center gap-2 text-base'>
                <ArchiveRestore className='h-4 w-4 text-primary' /> Restore
                Points
              </CardTitle>
              <CardDescription>
                Lihat, unduh, dan restore bundle backup terenkripsi per
                perangkat.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <Button asChild variant='outline'>
                <Link to='/ai-ops/restore-points'>Buka Restore Points</Link>
              </Button>
            </CardContent>
          </Card>
        </div>
        <p className='text-xs text-muted-foreground'>
          Chat AI Ops akan tersedia pada milestone M4 setelah pipeline approval
          stabil.
        </p>
      </Main>
    </>
  )
}
