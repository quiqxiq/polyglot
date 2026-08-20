import { Link } from '@tanstack/react-router'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  BookOpen,
  FileText,
  MessagesSquare,
  Network,
  Terminal,
  Ticket,
} from 'lucide-react'

export function QuickActions() {
  const actions = [
    {
      title: 'Generate Voucher',
      desc: 'Buat batch voucher hotspot baru',
      url: '/hotspot',
      search: { tab: 'users' as const },
      icon: Ticket,
      color: 'bg-emerald-500/10 text-emerald-600 hover:bg-emerald-500/20',
    },
    {
      title: 'Tambah PPPoE',
      desc: 'Daftarkan akun pelanggan baru',
      url: '/ppp',
      search: { tab: 'secrets' as const },
      icon: Network,
      color: 'bg-purple-500/10 text-purple-600 hover:bg-purple-500/20',
    },
    {
      title: 'Live Chat WhatsApp',
      desc: 'Lihat tiket & eskalasi pelanggan',
      url: '/chats',
      icon: MessagesSquare,
      color: 'bg-blue-500/10 text-blue-600 hover:bg-blue-500/20',
    },
    {
      title: 'Router Diagnostics',
      desc: 'Web terminal & ping test',
      url: '/devices',
      icon: Terminal,
      color: 'bg-amber-500/10 text-amber-600 hover:bg-amber-500/20',
    },
    {
      title: 'Laporan Penjualan',
      desc: 'Rekap omset & cetak laporan',
      url: '/reports',
      icon: FileText,
      color: 'bg-indigo-500/10 text-indigo-600 hover:bg-indigo-500/20',
    },
    {
      title: 'Knowledge Base AI',
      desc: 'Atur data konteks chatbot AI',
      url: '/knowledge',
      icon: BookOpen,
      color: 'bg-rose-500/10 text-rose-600 hover:bg-rose-500/20',
    },
  ]

  return (
    <Card className='col-span-3 shadow-xs'>
      <CardHeader className='pb-3'>
        <CardTitle className='text-base font-semibold'>Aksi Cepat & Navigasi</CardTitle>
        <CardDescription>Pintasan instan ke modul operasional harian</CardDescription>
      </CardHeader>
      <CardContent className='pt-1'>
        <div className='grid grid-cols-1 gap-2 sm:grid-cols-2'>
          {actions.map((act) => {
            const Icon = act.icon
            return (
              <Link
                key={act.title}
                to={act.url}
                className='group flex items-start gap-2.5 rounded-lg border p-2.5 transition-all hover:border-primary/50 hover:bg-muted/40'
              >
                <div
                  className={`flex size-8 shrink-0 items-center justify-center rounded-md transition-colors ${act.color}`}
                >
                  <Icon className='size-4' />
                </div>
                <div className='min-w-0 flex-1'>
                  <p className='text-xs font-semibold text-foreground group-hover:text-primary transition-colors'>
                    {act.title}
                  </p>
                  <p className='text-[11px] text-muted-foreground line-clamp-1'>
                    {act.desc}
                  </p>
                </div>
              </Link>
            )
          })}
        </div>
      </CardContent>
    </Card>
  )
}
