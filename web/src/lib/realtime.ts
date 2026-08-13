// Status koneksi SSE realtime WhatsApp — dipakai useWARealtimeStream untuk
// menghasilkan state, dan SSEIndicator untuk menampilkannya. Ditaruh di lib
// supaya komponen shared tidak bergantung ke modul feature.
export type WARealtimeStatus = 'connecting' | 'open' | 'reconnecting' | 'closed'
