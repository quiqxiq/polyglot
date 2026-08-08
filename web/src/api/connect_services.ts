import { createClient } from '@connectrpc/connect';
import { connectTransport } from './connect_client';

// ConnectRPC service descriptors for polyglot.v1.DeviceService and polyglot.v1.MikhmonService
export const DeviceService = {
  typeName: 'polyglot.v1.DeviceService',
  methods: [
    {
      name: 'ListDevices',
      I: class ListDevicesRequest {},
      O: class ListDevicesResponse {},
      kind: 'unary',
    },
    {
      name: 'GetDevice',
      I: class GetDeviceRequest {},
      O: class GetDeviceResponse {},
      kind: 'unary',
    },
    {
      name: 'UpdateDevice',
      I: class UpdateDeviceRequest {},
      O: class UpdateDeviceResponse {},
      kind: 'unary',
    },
    {
      name: 'StreamDeviceStatus',
      I: class StreamDeviceStatusRequest {},
      O: class DeviceStatusFrame {},
      kind: 'serverStreaming',
    },
    {
      name: 'StreamTerminal',
      I: class TerminalFrame {},
      O: class TerminalFrame {},
      kind: 'bidiStreaming',
    },
  ],
} as const;

export const MikhmonService = {
  typeName: 'polyglot.v1.MikhmonService',
  methods: [
    {
      name: 'GetDashboard',
      I: class GetMikhmonDashboardRequest {},
      O: class GetMikhmonDashboardResponse {},
      kind: 'unary',
    },
    {
      name: 'ListProfiles',
      I: class ListMikhmonProfilesRequest {},
      O: class ListMikhmonProfilesResponse {},
      kind: 'unary',
    },
    {
      name: 'ListUsers',
      I: class ListMikhmonUsersRequest {},
      O: class ListMikhmonUsersResponse {},
      kind: 'unary',
    },
    {
      name: 'ListActiveSessions',
      I: class ListMikhmonActiveSessionsRequest {},
      O: class ListMikhmonActiveSessionsResponse {},
      kind: 'unary',
    },
    {
      name: 'KickActiveSession',
      I: class KickMikhmonSessionRequest {},
      O: class KickMikhmonSessionResponse {},
      kind: 'unary',
    },
    {
      name: 'GenerateVouchers',
      I: class GenerateVouchersRequest {},
      O: class GenerateVouchersResponse {},
      kind: 'unary',
    },
  ],
} as const;

export const deviceConnectClient = createClient(DeviceService as any, connectTransport);
export const mikhmonConnectClient = createClient(MikhmonService as any, connectTransport);

export interface TerminalFrameMessage {
  deviceId?: string;
  inputData?: Uint8Array;
  outputData?: Uint8Array;
  cols?: number;
  rows?: number;
}
