import { createConnectTransport } from '@connectrpc/connect-web';

// ConnectRPC transport configured for NetOps Engine backend (:8080)
export const connectTransport = createConnectTransport({
  baseUrl: 'http://localhost:8080',
});
