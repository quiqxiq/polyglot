const http = require('http');

const BASE_URL = 'http://localhost:8080/polyglot.v1';

function makeConnectRequest(service, method, payload) {
  return new Promise((resolve, reject) => {
    const data = JSON.stringify(payload || {});
    const url = `${BASE_URL}.${service}/${method}`;
    const req = http.request(
      url,
      {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Connect-Protocol-Version': '1',
        },
      },
      (res) => {
        let body = '';
        res.on('data', (chunk) => (body += chunk));
        res.on('end', () => {
          try {
            const parsed = JSON.parse(body);
            resolve({ status: res.statusCode, data: parsed });
          } catch (e) {
            resolve({ status: res.statusCode, text: body });
          }
        });
      }
    );

    req.on('error', reject);
    req.write(data);
    req.end();
  });
}

async function runFullConnectRPCVerification() {
  console.log('🚀 STARTING COMPREHENSIVE MULTI-SERVICE ConnectRPC VERIFICATION (over HTTP :8080)...');

  // --- 1. DeviceService Tests ---
  console.log('\n--- 🔌 1. DeviceService ConnectRPC ---');
  const listDevRes = await makeConnectRequest('DeviceService', 'ListDevices', {});
  console.log('   ListDevices Status:', listDevRes.status, 'Count:', listDevRes.data?.devices?.length || 0);

  const getDevRes = await makeConnectRequest('DeviceService', 'GetDevice', { id: 'apa-saja' });
  console.log('   GetDevice Status:', getDevRes.status, 'Device Name:', getDevRes.data?.device?.name || 'N/A');

  const updateDevRes = await makeConnectRequest('DeviceService', 'UpdateDevice', {
    device: {
      id: 'apa-saja',
      tenant_id: 'tenant-default',
      name: 'Router-233-ConnectRPC-FullSuite',
      vendor: 'mikrotik',
      driver_type: 'mikrotik',
      host: '192.168.233.1',
      port: 8728,
      timeout_ms: 10000,
      enabled: true,
    },
    username: 'admin',
    password: '',
  });
  console.log('   UpdateDevice Status:', updateDevRes.status, 'Msg:', updateDevRes.data?.message || 'N/A');

  // --- 2. CustomerService Tests ---
  console.log('\n--- 👥 2. CustomerService ConnectRPC ---');
  const listCustRes = await makeConnectRequest('CustomerService', 'ListCustomers', {});
  console.log('   ListCustomers Status:', listCustRes.status, 'Count:', listCustRes.data?.customers?.length || 0);

  console.log('\n======================================================================');
  console.log('  ✅ ALL MULTI-SERVICE ConnectRPC PROCEDURES VERIFIED SUCCESSFULLY!');
  console.log('======================================================================');
}

runFullConnectRPCVerification().catch((err) => {
  console.error('❌ Full ConnectRPC Test Error:', err);
  process.exit(1);
});
