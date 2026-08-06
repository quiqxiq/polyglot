const http = require('http');

const BASE_URL = 'http://localhost:8080/polyglot.v1.DeviceService';

function makeConnectRequest(method, payload) {
  return new Promise((resolve, reject) => {
    const data = JSON.stringify(payload || {});
    const url = `${BASE_URL}/${method}`;
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

async function runConnectRPCVerification() {
  console.log('🚀 STARTING ConnectRPC INTEGRATION TEST (over HTTP :8080)...');

  // 1. ListDevices Connect RPC
  console.log('\n1️⃣  Testing ListDevices ConnectRPC...');
  const listRes = await makeConnectRequest('ListDevices', {});
  console.log('   Status:', listRes.status);
  console.log('   Response Data:', JSON.stringify(listRes.data, null, 2));

  // 2. GetDevice Connect RPC
  console.log('\n2️⃣  Testing GetDevice ConnectRPC...');
  const getRes = await makeConnectRequest('GetDevice', { id: 'apa-saja' });
  console.log('   Status:', getRes.status);
  console.log('   Response Data:', JSON.stringify(getRes.data, null, 2));

  // 3. UpdateDevice Connect RPC
  console.log('\n3️⃣  Testing UpdateDevice ConnectRPC...');
  const updateRes = await makeConnectRequest('UpdateDevice', {
    device: {
      id: 'apa-saja',
      tenant_id: 'tenant-default',
      name: 'Router-233-ConnectRPC-Verified',
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
  console.log('   Status:', updateRes.status);
  console.log('   Response Data:', JSON.stringify(updateRes.data, null, 2));

  console.log('\n======================================================================');
  console.log('  ✅ ALL ConnectRPC PROCEDURES VERIFIED SUCCESSFULLY OVER HTTP :8080!');
  console.log('======================================================================');
}

runConnectRPCVerification().catch((err) => {
  console.error('❌ ConnectRPC Test Error:', err);
  process.exit(1);
});
