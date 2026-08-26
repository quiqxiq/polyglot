// Katalog resource & action yang dikenal sistem — dipakai untuk Permission Matrix & RBAC
// Harus sinkron dengan internal/adapter/auth/procedure_permissions.go.

export interface PermissionDefinition {
  id: string // Format: "resource:action" (contoh: "device:read")
  resource: string
  action: string
  label: string
  description: string
}

export interface ModuleGroup {
  id: string
  label: string
  description: string
  permissions: PermissionDefinition[]
}

export const RBAC_MODULE_GROUPS: ModuleGroup[] = [
  {
    id: 'devices',
    label: 'Devices & Routers',
    description: 'Manajemen router MikroTik, status koneksi, terminal, dan diagnostik',
    permissions: [
      {
        id: 'device:read',
        resource: 'device',
        action: 'read',
        label: 'View Devices',
        description: 'Melihat daftar perangkat, detail status, dan traffic interface',
      },
      {
        id: 'device:manage',
        resource: 'device',
        action: 'manage',
        label: 'Manage Devices',
        description: 'Menambah, mengubah, dan menghapus inventaris perangkat',
      },
      {
        id: 'device:diagnostic:exec',
        resource: 'device',
        action: 'diagnostic:exec',
        label: 'Diagnostic & Ping',
        description: 'Test connection dan live ping diagnostic ke router',
      },
      {
        id: 'device:terminal:exec',
        resource: 'device',
        action: 'terminal:exec',
        label: 'Web Terminal',
        description: 'Akses interaktif SSH Web Terminal MikroTik',
      },
    ],
  },
  {
    id: 'hotspot',
    label: 'Hotspot (Mikhmon)',
    description: 'Manajemen voucher, user, profile, host, DHCP, dan sesi aktif hotspot',
    permissions: [
      {
        id: 'hotspot:voucher:generate',
        resource: 'hotspot',
        action: 'voucher:generate',
        label: 'Generate & Print Vouchers',
        description: 'Generate batch voucher, cetak voucher template, dan cek status',
      },
      {
        id: 'hotspot:user:read',
        resource: 'hotspot',
        action: 'user:read',
        label: 'View Hotspot Users',
        description: 'Melihat daftar akun voucher/user hotspot',
      },
      {
        id: 'hotspot:user:write',
        resource: 'hotspot',
        action: 'user:write',
        label: 'Manage Hotspot Users',
        description: 'Menambah, mengedit, reset counter, dan menghapus user hotspot',
      },
      {
        id: 'hotspot:profile:read',
        resource: 'hotspot',
        action: 'profile:read',
        label: 'View Profiles',
        description: 'Melihat daftar profile paket kecepatan dan masa aktif',
      },
      {
        id: 'hotspot:profile:write',
        resource: 'hotspot',
        action: 'profile:write',
        label: 'Manage Profiles',
        description: 'Menambah, mengedit, dan menghapus profile paket hotspot',
      },
      {
        id: 'hotspot:active:read',
        resource: 'hotspot',
        action: 'active:read',
        label: 'View Active Sessions',
        description: 'Melihat user online/sesi aktif dan traffic realtime',
      },
      {
        id: 'hotspot:active:kick',
        resource: 'hotspot',
        action: 'active:kick',
        label: 'Kick Active Sessions',
        description: 'Memutuskan / kick koneksi user hotspot yang sedang aktif',
      },
      {
        id: 'hotspot:dhcp:read',
        resource: 'hotspot',
        action: 'dhcp:read',
        label: 'View DHCP Leases',
        description: 'Melihat daftar alokasi IP DHCP server',
      },
      {
        id: 'hotspot:dhcp:manage',
        resource: 'hotspot',
        action: 'dhcp:manage',
        label: 'Manage DHCP Leases',
        description: 'Membuat static binding atau memblokir lease DHCP',
      },
      {
        id: 'hotspot:host:read',
        resource: 'hotspot',
        action: 'host:read',
        label: 'View Hosts',
        description: 'Melihat tabel host yang terdeteksi di jaringan hotspot',
      },
      {
        id: 'hotspot:host:manage',
        resource: 'hotspot',
        action: 'host:manage',
        label: 'Manage Hosts',
        description: 'Menghapus host dari tabel hotspot router',
      },
      {
        id: 'hotspot:binding:read',
        resource: 'hotspot',
        action: 'binding:read',
        label: 'View IP Bindings & Cookies',
        description: 'Melihat IP binding (bypass/blocked) dan cookie aktif',
      },
      {
        id: 'hotspot:binding:manage',
        resource: 'hotspot',
        action: 'binding:manage',
        label: 'Manage IP Bindings & Cookies',
        description: 'Membuat bypass IP, regular binding, dan hapus cookie',
      },
      {
        id: 'hotspot:report:read',
        resource: 'hotspot',
        action: 'report:read',
        label: 'View Sales Reports',
        description: 'Melihat ringkasan laporan penjualan voucher',
      },
      {
        id: 'hotspot:report:manage',
        resource: 'hotspot',
        action: 'report:manage',
        label: 'Manage Sales Reports',
        description: 'Menghapus data rekaman laporan penjualan voucher',
      },
      {
        id: 'hotspot:expire:read',
        resource: 'hotspot',
        action: 'expire:read',
        label: 'View Expire Monitor',
        description: 'Melihat status service monitor masa aktif voucher',
      },
      {
        id: 'hotspot:expire:manage',
        resource: 'hotspot',
        action: 'expire:manage',
        label: 'Manage Expire Monitor',
        description: 'Setup, aktifkan, dan nonaktifkan script expire monitor di router',
      },
      {
        id: 'hotspot:template:read',
        resource: 'hotspot',
        action: 'template:read',
        label: 'View Print Templates',
        description: 'Melihat preview template cetak voucher',
      },
      {
        id: 'hotspot:read',
        resource: 'hotspot',
        action: 'read',
        label: 'General Hotspot Telemetry',
        description: 'Akses stream telemetry resource, ethernet traffic, dan pool',
      },
    ],
  },
  {
    id: 'ppp',
    label: 'PPP (PPPoE)',
    description: 'Manajemen akun rahasia PPPoE, profil kecepatan, sesi aktif, dan kick',
    permissions: [
      {
        id: 'ppp:secret:read',
        resource: 'ppp',
        action: 'secret:read',
        label: 'View PPP Secrets',
        description: 'Melihat daftar akun secret PPPoE pelanggan',
      },
      {
        id: 'ppp:secret:write',
        resource: 'ppp',
        action: 'secret:write',
        label: 'Manage PPP Secrets',
        description: 'Menambah, mengubah password, disable/enable, dan menghapus akun PPPoE',
      },
      {
        id: 'ppp:profile:read',
        resource: 'ppp',
        action: 'profile:read',
        label: 'View PPP Profiles',
        description: 'Melihat profil paket kecepatan PPPoE',
      },
      {
        id: 'ppp:profile:write',
        resource: 'ppp',
        action: 'profile:write',
        label: 'Manage PPP Profiles',
        description: 'Menambah, mengedit rate limit, dan menghapus profil PPPoE',
      },
      {
        id: 'ppp:active:read',
        resource: 'ppp',
        action: 'active:read',
        label: 'View PPP Active Sessions',
        description: 'Melihat statistik koneksi pelanggan PPPoE yang sedang online/offline',
      },
      {
        id: 'ppp:active:kick',
        resource: 'ppp',
        action: 'active:kick',
        label: 'Kick PPP Active Sessions',
        description: 'Memutuskan / kick koneksi pelanggan PPPoE yang sedang online',
      },
    ],
  },
  {
    id: 'customers',
    label: 'Customer & CRM',
    description: 'Data pelanggan ISP, koordinat ODP, dan kode akses portal',
    permissions: [
      {
        id: 'customer:read',
        resource: 'customer',
        action: 'read',
        label: 'View Customers',
        description: 'Melihat database pelanggan, detail kontak, dan geolokasi',
      },
      {
        id: 'customer:manage',
        resource: 'customer',
        action: 'manage',
        label: 'Manage Customers',
        description: 'Menambah, mengubah, dan menghapus data pelanggan',
      },
    ],
  },
  {
    id: 'registrations',
    label: 'Registrasi & Pasang Baru',
    description: 'Pendaftaran calon pelanggan, penjadwalan, instalasi fisik teknisi, dan konversi',
    permissions: [
      {
        id: 'registration:read',
        resource: 'registration',
        action: 'read',
        label: 'View Registrations',
        description: 'Melihat daftar calon pelanggan dan status permohonan pasang',
      },
      {
        id: 'registration:manage',
        resource: 'registration',
        action: 'manage',
        label: 'Manage Registrations',
        description: 'Approve pendaftaran, jadwalkan teknisi, reject/cancel, dan konversi ke pelanggan aktif',
      },
      {
        id: 'registration:install',
        resource: 'registration',
        action: 'install',
        label: 'Mark Installed (Teknisi)',
        description: 'Menandai instalasi fisik telah selesai dan mencatat hasil redaman',
      },
    ],
  },
  {
    id: 'billing',
    label: 'Billing & Invoicing',
    description: 'Faktur tagihan bulanan, kasir cepat, paket layanan, dan lifecycle langganan',
    permissions: [
      {
        id: 'billing:read',
        resource: 'billing',
        action: 'read',
        label: 'View Billing',
        description: 'Melihat faktur, langganan aktif/isolir, paket layanan, dan lookup kasir',
      },
      {
        id: 'billing:manage',
        resource: 'billing',
        action: 'manage',
        label: 'Manage Billing & Cashier',
        description: 'Proses pembayaran kasir, ganti paket, suspend/resume/terminate, dan generate tagihan',
      },
    ],
  },
  {
    id: 'cashbook',
    label: 'Buku Kas & Keuangan',
    description: 'Pencatatan kasir fisik, rekening bank, mutasi pemasukan/pengeluaran, dan saldo kas',
    permissions: [
      {
        id: 'cashbook:read',
        resource: 'cashbook',
        action: 'read',
        label: 'View Cashbook',
        description: 'Melihat rekening kas, kategori transaksi, mutasi jurnal kas, dan saldo',
      },
      {
        id: 'cashbook:manage',
        resource: 'cashbook',
        action: 'manage',
        label: 'Manage Cashbook',
        description: 'Menambah/mengubah rekening, kategori, dan mencatat transaksi kas masuk/keluar',
      },
    ],
  },
  {
    id: 'notifications',
    label: 'WhatsApp Notifications',
    description: 'Template pesan WhatsApp tagihan/kuitansi, antrean notifikasi, dan monitoring kirim',
    permissions: [
      {
        id: 'notification:read',
        resource: 'notification',
        action: 'read',
        label: 'View Notifications',
        description: 'Melihat template pesan, antrean notifikasi terkirim/gagal, dan pending count',
      },
      {
        id: 'notification:manage',
        resource: 'notification',
        action: 'manage',
        label: 'Manage Notifications',
        description: 'Menyimpan template pesan, trigger uji coba kirim, dan update antrean',
      },
    ],
  },
  {
    id: 'reports',
    label: 'Financial & Operational Reports',
    description: 'Laporan pendapatan, piutang, pengeluaran, dan performa keuangan berkala',
    permissions: [
      {
        id: 'report:read',
        resource: 'report',
        action: 'read',
        label: 'View Reports',
        description: 'Melihat laporan finansial harian, bulanan, dan tahunan',
      },
      {
        id: 'report:manage',
        resource: 'report',
        action: 'manage',
        label: 'Refresh Reports',
        description: 'Trigger rebuild kalkulasi snapshot laporan finansial manual',
      },
    ],
  },
  {
    id: 'ispadmin',
    label: 'ISP Migration & Reconcile',
    description: 'Impor/ekspor data pelanggan CSV/Excel, tarik akun router, dan rekonsiliasi drift',
    permissions: [
      {
        id: 'ispadmin:manage',
        resource: 'ispadmin',
        action: 'manage',
        label: 'Manage ISP Admin Operations',
        description: 'Impor file, ekspor database, tarik akun router, dan periksa deviasi router vs database',
      },
    ],
  },
  {
    id: 'whatsapp',
    label: 'WhatsApp & Live Chat',
    description: 'Integrasi WhatsApp, sesi pairing, chatbot, dan ambil alih agen',
    permissions: [
      {
        id: 'whatsapp:read',
        resource: 'whatsapp',
        action: 'read',
        label: 'View WhatsApp',
        description: 'Melihat sesi WA terhubung dan riwayat pesan chat',
      },
      {
        id: 'whatsapp:message',
        resource: 'whatsapp',
        action: 'message',
        label: 'Send Messages',
        description: 'Mengirimkan pesan teks/balasan WhatsApp manual',
      },
      {
        id: 'whatsapp:write',
        resource: 'whatsapp',
        action: 'write',
        label: 'Chat Actions',
        description: 'Tandai sudah dibaca dan toggle bot per chat',
      },
      {
        id: 'whatsapp:manage',
        resource: 'whatsapp',
        action: 'manage',
        label: 'Manage Sessions',
        description: 'Scan QR code, pairing nomor, logout, dan toggle sesi bot',
      },
      {
        id: 'conversation:read',
        resource: 'conversation',
        action: 'read',
        label: 'View Conversations',
        description: 'Melihat percakapan bot dan ringkasan konteks',
      },
      {
        id: 'conversation:write',
        resource: 'conversation',
        action: 'write',
        label: 'Agent Takeover',
        description: 'Ambil alih percakapan dari AI bot dan kembalikan ke bot',
      },
    ],
  },
  {
    id: 'skills',
    label: 'Skills, SOP & AI LLM',
    description: 'Manajemen modular skills, SOP layanan, dan konfigurasi model LLM',
    permissions: [
      {
        id: 'skill:read',
        resource: 'skill',
        action: 'read',
        label: 'View Skills & Prompts',
        description: 'Melihat daftar modular skills, file SOP, dan base system prompt',
      },
      {
        id: 'skill:manage',
        resource: 'skill',
        action: 'manage',
        label: 'Manage Skills & Prompts',
        description: 'Membuat, mengedit berkas SOP, menghapus skill, dan sinkronisasi ke disk',
      },
      {
        id: 'llmconfig:read',
        resource: 'llmconfig',
        action: 'read',
        label: 'View LLM Config',
        description: 'Melihat daftar provider AI dan model aktif',
      },
      {
        id: 'llmconfig:manage',
        resource: 'llmconfig',
        action: 'manage',
        label: 'Manage LLM Config',
        description: 'Mengubah API key, endpoint provider, dan ganti model AI',
      },
      {
        id: 'technician:read',
        resource: 'technician',
        action: 'read',
        label: 'View Technicians',
        description: 'Melihat daftar kontak teknisi lapangan untuk eskalasi bot',
      },
      {
        id: 'technician:manage',
        resource: 'technician',
        action: 'manage',
        label: 'Manage Technicians',
        description: 'Menambah, mengedit, dan menghapus kontak teknisi',
      },
    ],
  },
  {
    id: 'settings',
    label: 'System Settings',
    description: 'Pengaturan umum sistem, parameter anti-spam bot, dan preferensi aplikasi',
    permissions: [
      {
        id: 'setting:read',
        resource: 'setting',
        action: 'read',
        label: 'View Settings',
        description: 'Melihat parameter konfigurasi sistem dan bot',
      },
      {
        id: 'setting:manage',
        resource: 'setting',
        action: 'manage',
        label: 'Manage Settings',
        description: 'Mengubah konfigurasi sistem, batch update, dan anti-spam bot settings',
      },
    ],
  },
  {
    id: 'system_admin',
    label: 'Users & RBAC Administration',
    description: 'Manajemen akun staf, role-based access control matrix, dan probe',
    permissions: [
      {
        id: 'user:read',
        resource: 'user',
        action: 'read',
        label: 'View Users',
        description: 'Melihat daftar akun staf dan status aktif',
      },
      {
        id: 'user:manage',
        resource: 'user',
        action: 'manage',
        label: 'Manage Users',
        description: 'Membuat akun staf, reset password, dan blokir akun',
      },
      {
        id: 'rbac:manage',
        resource: 'rbac',
        action: 'manage',
        label: 'Manage RBAC',
        description: 'Membuat role baru, konfigurasi matrix izin, dan assign role',
      },
      {
        id: 'probe:read',
        resource: 'probe',
        action: 'read',
        label: 'View Probes',
        description: 'Melihat laporan status dan telemetri remote probe',
      },
    ],
  },
  {
    id: 'logs',
    label: 'Router & System Logs',
    description: 'Pemantauan live stream log MikroTik, Hotspot, dan PPP',
    permissions: [
      {
        id: 'log:read',
        resource: 'log',
        action: 'read',
        label: 'View Router Logs',
        description: 'Melihat live stream log MikroTik (All, Hotspot, dan PPP)',
      },
    ],
  },
]

export const ALL_PERMISSION_IDS = RBAC_MODULE_GROUPS.flatMap((g) =>
  g.permissions.map((p) => p.id)
)

export const RBAC_RESOURCES = [
  'device',
  'customer',
  'registration',
  'billing',
  'cashbook',
  'notification',
  'report',
  'ispadmin',
  'conversation',
  'skill',
  'whatsapp',
  'user',
  'rbac',
  'llmconfig',
  'technician',
  'probe',
  'hotspot',
  'ppp',
  'setting',
  'log',
] as const

export const RBAC_ACTIONS = [
  'read',
  'write',
  'manage',
  'command',
  'message',
  'install',
] as const

export const ALL = '*'

export function composeObject(resource: string, action: string): string {
  if (resource === ALL) return '.*'
  if (action === ALL) return `${resource}:.*`
  return `${resource}:${action}`
}

export function splitObject(obj: string): { resource: string; action: string } {
  if (obj === '.*') return { resource: ALL, action: ALL }
  const idx = obj.indexOf(':')
  if (idx === -1) return { resource: obj, action: '' }
  const action = obj.slice(idx + 1)
  return { resource: obj.slice(0, idx), action: action === '*' ? ALL : action }
}
