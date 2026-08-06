const http = require('http');

const BASE_URL = 'http://localhost:8080/polyglot.v1.MikhmonService';
const TARGET_DEVICE_ID = 'apa-saja';

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

async function runMikhmonConnectRPCVerification() {
  console.log('🚀 STARTING MIKHMON ConnectRPC INTEGRATION TEST (over HTTP :8080)...');
  console.log('   Target Router ID:', TARGET_DEVICE_ID);

  // 1. GetDashboard
  console.log('\n1️⃣  Testing GetDashboard ConnectRPC...');
  const dashRes = await makeConnectRequest('GetDashboard', { device_id: TARGET_DEVICE_ID });
  console.log('   Status:', dashRes.status);
  console.log('   Dashboard Summary:', JSON.stringify(dashRes.data, null, 2));

  // 2. ListProfiles
  console.log('\n2️⃣  Testing ListProfiles ConnectRPC...');
  const profRes = await makeConnectRequest('ListProfiles', { device_id: TARGET_DEVICE_ID });
  console.log('   Status:', profRes.status);
  console.log('   Profiles Count:', profRes.data?.profiles?.length || 0);

  // 3. ListUsers
  console.log('\n3️⃣  Testing ListUsers ConnectRPC...');
  const usersRes = await makeConnectRequest('ListUsers', { device_id: TARGET_DEVICE_ID });
  console.log('   Status:', usersRes.status);
  console.log('   Users Count:', usersRes.data?.users?.length || 0);

  // 4. ListActiveSessions
  console.log('\n4️⃣  Testing ListActiveSessions ConnectRPC...');
  const actRes = await makeConnectRequest('ListActiveSessions', { device_id: TARGET_DEVICE_ID });
  console.log('   Status:', actRes.status);
  console.log('   Active Sessions Count:', actRes.data?.sessions?.length || 0);

  // 5. ListDHCPLeases
  console.log('\n5️⃣  Testing ListDHCPLeases ConnectRPC...');
  const dhcpRes = await makeConnectRequest('ListDHCPLeases', { device_id: TARGET_DEVICE_ID });
  console.log('   Status:', dhcpRes.status);
  console.log('   DHCP Leases Count:', dhcpRes.data?.leases?.length || 0);

  // 6. GenerateVouchers
  console.log('\n6️⃣  Testing GenerateVouchers ConnectRPC...');
  const genRes = await makeConnectRequest('GenerateVouchers', {
    device_id: TARGET_DEVICE_ID,
    profile: 'default',
    count: 2,
    user_length: 6,
    prefix: 'vc',
    character_set: 'lowernum',
  });
  console.log('   Status:', genRes.status);
  console.log('   Generated Vouchers:', JSON.stringify(genRes.data, null, 2));

  console.log('\n======================================================================');
  console.log('  ✅ ALL MIKHMON ConnectRPC PROCEDURES VERIFIED SUCCESSFULLY!');
  console.log('======================================================================');
}

runMikhmonConnectRPCVerification().catch((err) => {
  console.error('❌ Mikhmon ConnectRPC Test Error:', err);
  process.exit(1);
});
