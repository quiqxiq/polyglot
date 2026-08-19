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
    description: 'Manajemen router MikroTik, status koneksi, dan diagnostik',
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
        id: 'device:command',
        resource: 'device',
        action: 'command',
        label: 'Execute & Terminal',
        description: 'Akses Web Terminal, test connection, dan live ping',
      },
    ],
  },
  {
    id: 'hotspot',
    label: 'Hotspot (Mikhmon)',
    description: 'Manajemen voucher, user profile, dan sesi aktif hotspot',
    permissions: [
      {
        id: 'hotspot:read',
        resource: 'hotspot',
        action: 'read',
        label: 'View Hotspot',
        description: 'Melihat user, profile, host, DHCP leases, dan sesi aktif',
      },
      {
        id: 'hotspot:manage',
        resource: 'hotspot',
        action: 'manage',
        label: 'Manage Hotspot',
        description: 'Generate voucher, edit profile/user, kick sesi, dan block DHCP',
      },
    ],
  },
  {
    id: 'ppp',
    label: 'PPP (PPPoE)',
    description: 'Manajemen akun rahasia PPPoE, sesi aktif, dan kick pelanggan',
    permissions: [
      {
        id: 'ppp:read',
        resource: 'ppp',
        action: 'read',
        label: 'View PPP',
        description: 'Melihat daftar secret, statistik sesi aktif, dan inactive user',
      },
      {
        id: 'ppp:manage',
        resource: 'ppp',
        action: 'manage',
        label: 'Manage PPP',
        description: 'Menambah/mengedit secret, reset password, dan kick user PPP',
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
    id: 'knowledge',
    label: 'Knowledge Base & AI LLM',
    description: 'Dokumen panduan, vektor embeddings, dan konfigurasi model LLM',
    permissions: [
      {
        id: 'knowledge:read',
        resource: 'knowledge',
        action: 'read',
        label: 'View Knowledge',
        description: 'Melihat dokumen pengetahuan dan status embedding',
      },
      {
        id: 'knowledge:write',
        resource: 'knowledge',
        action: 'write',
        label: 'Manage Knowledge',
        description: 'Upload, edit konten, dan hapus dokumen pengetahuan',
      },
      {
        id: 'knowledge:embed',
        resource: 'knowledge',
        action: 'embed',
        label: 'Embed Documents',
        description: 'Memicu embedding dokumen ke vector database',
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
    id: 'billing_customer',
    label: 'Billing & Customers',
    description: 'Data pelanggan, langganan internet, invoice, dan pembayaran',
    permissions: [
      {
        id: 'customer:read',
        resource: 'customer',
        action: 'read',
        label: 'View Customers',
        description: 'Melihat data pelanggan dan profil kontak',
      },
      {
        id: 'billing:read',
        resource: 'billing',
        action: 'read',
        label: 'View Invoices',
        description: 'Melihat tagihan, status invoice, dan paket langganan',
      },
      {
        id: 'billing:write',
        resource: 'billing',
        action: 'write',
        label: 'Manage Invoices',
        description: 'Membuat invoice, memproses pembayaran, dan ubah langganan',
      },
    ],
  },
  {
    id: 'system_admin',
    label: 'Users & System Administration',
    description: 'Manajemen akun staf, role-based access control, dan probe',
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
]

export const ALL_PERMISSION_IDS = RBAC_MODULE_GROUPS.flatMap((g) =>
  g.permissions.map((p) => p.id)
)

export const RBAC_RESOURCES = [
  'device',
  'customer',
  'conversation',
  'knowledge',
  'billing',
  'whatsapp',
  'user',
  'rbac',
  'llmconfig',
  'technician',
  'probe',
  'hotspot',
  'ppp',
] as const

export const RBAC_ACTIONS = [
  'read',
  'write',
  'manage',
  'command',
  'embed',
  'message',
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
