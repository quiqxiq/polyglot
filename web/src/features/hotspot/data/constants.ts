export const EXPIRE_MODES = [
  { value: '0', label: 'None / Disabled', desc: 'No expiration action' },
  { value: 'rem', label: 'Remove', desc: 'Hapus user saat expired' },
  { value: 'ntf', label: 'Notice', desc: 'Kunci user & beri pesan expired' },
  { value: 'remc', label: 'Remove & Record', desc: 'Hapus user & catat ke Laporan Penjualan' },
  { value: 'ntfc', label: 'Notice & Record', desc: 'Beri pesan & catat ke Laporan Penjualan' },
] as const

export const USER_MODES = [
  { value: 'vc', label: 'Username = Password (Voucher)' },
  { value: 'up', label: 'Username & Password (Member)' },
] as const

export const CHAR_SETS = [
  { value: 'mix', label: '5ab2c34d (Random Numbers & Lowercase)' },
  { value: 'mix1', label: '5AB2C34D (Random Numbers & Uppercase)' },
  { value: 'mix2', label: '5aB2c34D (Numbers, Lower & Uppercase)' },
  { value: 'num', label: '12345678 (Numbers Only)' },
  { value: 'lower', label: 'abcdefgh (Lowercase Only)' },
  { value: 'upper', label: 'ABCDEFGH (Uppercase Only)' },
] as const

export const TEMPLATE_LAYOUTS = [
  { value: 'default', label: 'Default Voucher (Normal Paper)' },
  { value: 'small', label: 'Small Voucher (Compact Paper)' },
  { value: 'thermal', label: 'Thermal Roll (58mm / 80mm POS Printer)' },
] as const
