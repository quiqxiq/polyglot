const http = require('http');

const BASE_URL = 'http://localhost:8080';
const DEVICE_ID = 'apa-saja';

// Helper function to send REST HTTP requests
function sendRequest(path, method, body = null) {
  return new Promise((resolve, reject) => {
    const url = new URL(BASE_URL + path);
    const options = {
      hostname: url.hostname,
      port: url.port,
      path: url.pathname + url.search,
      method: method,
      headers: {
        'Content-Type': 'application/json',
      },
    };

    const req = http.request(options, (res) => {
      let data = '';
      res.on('data', (chunk) => (data += chunk));
      res.on('end', () => {
        try {
          resolve({ status: res.statusCode, data: data ? JSON.parse(data) : null });
        } catch (e) {
          resolve({ status: res.statusCode, raw: data });
        }
      });
    });

    req.on('error', reject);
    if (body) {
      req.write(JSON.stringify(body));
    }
    req.end();
  });
}

// Function to listen to SSE Stream
function subscribeSSE(path, onFrame) {
  return new Promise((resolve, reject) => {
    const url = new URL(BASE_URL + path);
    const req = http.get(url, (res) => {
      console.log(`[SSE Connected] Endpoint: ${path} (Status: ${res.statusCode})`);
      let buffer = '';

      res.on('data', (chunk) => {
        buffer += chunk.toString();
        const lines = buffer.split('\n');
        buffer = lines.pop(); // keep last incomplete line

        let currentEvent = '';
        let currentData = '';

        for (const line of lines) {
          if (line.startsWith('event:')) {
            currentEvent = line.substring(6).trim();
          } else if (line.startsWith('data:')) {
            currentData = line.substring(5).trim();
            if (currentEvent && currentData) {
              onFrame(currentEvent, currentData);
              currentEvent = '';
              currentData = '';
            }
          }
        }
      });

      // Keep open for testing
      resolve({ req, res });
    });

    req.on('error', reject);
  });
}

async function runTest() {
  console.log('=== 🚀 STARTING BACKEND SSE WIRE STREAMING INTEGRATION TEST ===\n');

  // STEP 1: Update device credentials to 192.168.233.1, admin, r00t
  console.log('1️⃣  Configuring router target to 192.168.233.1 with credentials admin/r00t...');
  const initPayload = {
    id: DEVICE_ID,
    name: 'Router-Initial-233',
    vendor: 'mikrotik',
    driver_type: 'mikrotik',
    host: '192.168.233.1',
    port: 8728,
    timeout_ms: 10000,
    enabled: true,
    username: 'admin',
    password: 'r00t',
  };

  const initRes = await sendRequest(`/api/v1/devices/${DEVICE_ID}`, 'PUT', initPayload);
  console.log('   PUT Initial Response:', initRes.status, JSON.stringify(initRes.data));

  // STEP 2: Connect SSE Stream Client
  console.log('\n2️⃣  Subscribing to SSE Stream at /ws/devices/stream...');
  const receivedFrames = [];

  const sseConnection = await subscribeSSE('/ws/devices/stream', (event, dataStr) => {
    try {
      const parsed = JSON.parse(dataStr);
      console.log(`\n📩  [SSE FRAME RECEIVED] Event: "${event}"`);
      const targetItem = Array.isArray(parsed) ? parsed.find((i) => i.device && i.device.id === DEVICE_ID) : null;
      if (targetItem) {
        console.log(`   Device Name in SSE Frame: "${targetItem.device.name}"`);
        console.log(`   Host IP in SSE Frame:     "${targetItem.device.host}"`);
        console.log(`   Status in SSE Frame:      "${targetItem.test.status}"`);
        console.log(`   Message in SSE Frame:     "${targetItem.test.message}"`);
        if (targetItem.test.identity) console.log(`   Identity:                 "${targetItem.test.identity}"`);
        if (targetItem.test.version) console.log(`   ROS Version:              "${targetItem.test.version}"`);
        if (targetItem.test.board_name) console.log(`   Board Name:               "${targetItem.test.board_name}"`);
        if (targetItem.test.uptime) console.log(`   Uptime:                   "${targetItem.test.uptime}"`);
      }
      receivedFrames.push({ event, parsed, targetItem });
    } catch (e) {
      console.log(`   Raw Data: ${dataStr}`);
    }
  });

  // Wait 3 seconds for initial streaming frames to settle
  await new Promise((r) => setTimeout(r, 3000));

  // STEP 3: Execute REST PUT update while SSE stream is active
  console.log('\n3️⃣  Executing REST PUT update (Changing name to "Router-Live-Streaming-233-SUCCESS", password: "")...');
  const updatePayload = {
    id: DEVICE_ID,
    name: 'Router-Live-Streaming-233-SUCCESS',
    vendor: 'mikrotik',
    driver_type: 'mikrotik',
    host: '192.168.233.1',
    port: 8728,
    timeout_ms: 10000,
    enabled: true,
    username: 'admin',
    password: '', // Blank password -> must preserve r00t from vault
  };

  const updateRes = await sendRequest(`/api/v1/devices/${DEVICE_ID}`, 'PUT', updatePayload);
  console.log('   PUT Update Response Status:', updateRes.status);
  console.log('   PUT Update Response Body:  ', JSON.stringify(updateRes.data));

  // Wait 4 seconds to observe real-time SSE stream frames after update
  console.log('\n4️⃣  Observing SSE Stream frames post-update for 4 seconds...');
  await new Promise((r) => setTimeout(r, 4000));

  // Close SSE Connection
  sseConnection.req.destroy();
  console.log('\n5️⃣  Closing SSE connection.');

  // STEP 4: Final DB Verification via REST GET
  console.log('\n6️⃣  Verifying DB state via GET /api/v1/devices...');
  const getRes = await sendRequest('/api/v1/devices', 'GET');
  console.log('   GET Response:', JSON.stringify(getRes.data, null, 2));

  console.log('\n=== ✅ BACKEND SSE TEST COMPLETED SUCCESSFULLY ===');
}

runTest().catch((err) => {
  console.error('❌ Test failed with error:', err);
  process.exit(1);
});
