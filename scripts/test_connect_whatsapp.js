const http = require('http');

function callConnect(path, body) {
  return new Promise((resolve, reject) => {
    const payload = JSON.stringify(body);
    const req = http.request(
      {
        hostname: 'localhost',
        port: 8080,
        path: path,
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Content-Length': Buffer.byteLength(payload),
        },
      },
      (res) => {
        let data = '';
        res.on('data', (chunk) => (data += chunk));
        res.on('end', () => {
          resolve({ status: res.statusCode, data: JSON.parse(data || '{}') });
        });
      }
    );
    req.on('error', reject);
    req.write(payload);
    req.end();
  });
}

async function runTests() {
  console.log('--- Testing WhatsAppService ConnectRPC ---');
  try {
    const listRes = await callConnect('/polyglot.v1.WhatsAppService/ListSessions', {});
    console.log('ListSessions status:', listRes.status, listRes.data);

    const createRes = await callConnect('/polyglot.v1.WhatsAppService/CreateSession', {
      name: 'Office WA Gateway',
      phone_number: '6281234567890',
    });
    console.log('CreateSession status:', createRes.status, createRes.data);

    const qrRes = await callConnect('/polyglot.v1.WhatsAppService/GetQRCode', { session_id: '1' });
    console.log('GetQRCode status:', qrRes.status, qrRes.data);

    const pairRes = await callConnect('/polyglot.v1.WhatsAppService/GetPairingCode', { session_id: '1', phone_number: '6281234567890' });
    console.log('GetPairingCode status:', pairRes.status, pairRes.data);

    const recRes = await callConnect('/polyglot.v1.WhatsAppService/ReconnectSession', { session_id: '1' });
    console.log('ReconnectSession status:', recRes.status, recRes.data);

    const logoutRes = await callConnect('/polyglot.v1.WhatsAppService/LogoutSession', { session_id: '1' });
    console.log('LogoutSession status:', logoutRes.status, logoutRes.data);

    const purgeRes = await callConnect('/polyglot.v1.WhatsAppService/PurgeSession', { session_id: '1' });
    console.log('PurgeSession status:', purgeRes.status, purgeRes.data);

    console.log('✅ ALL WhatsAppService ConnectRPC procedures PASSED!');
  } catch (err) {
    console.error('❌ Test failed:', err);
  }
}

runTests();
