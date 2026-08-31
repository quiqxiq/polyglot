import type { Device } from '@/gen/v1/device_pb'

export type DevicesDialogType =
  | 'add'
  | 'edit'
  | 'delete'
  | 'test'
  | 'terminal'
  | 'ping-settings'
  | 'ping-analytics'
  | 'isolation'
  | 'webhook-scripts'


export type ViewMode = 'card' | 'table'

export type VendorFilter = 'all' | 'mikrotik' | 'cisco' | 'huawei' | 'genieacs'

export type SortOrder = 'asc' | 'desc'

export type TimePreset =
  | '1h'
  | '6h'
  | '12h'
  | '24h'
  | 'today'
  | '3d'
  | '7d'
  | '15d'
  | '30d'
  | 'custom'


export type BucketInterval = '' | 'raw' | '1m' | '5m' | '1h'

export interface DeviceInterfaceItem {
  id?: string
  name: string
  type: string
  running: boolean
  disabled: boolean
}

export interface PingDataPoint {
  ms: number
  alive: boolean
  timestamp?: number
}

export interface TrafficDataPoint {
  time: number
  rx: number
  tx: number
}

export interface PingSummaryStats {
  minRtt: number
  avgRtt: number
  maxRtt: number
  packetLossPct: number
  totalSamples: number
}

export interface DeviceCardProps {
  device: Device
}

