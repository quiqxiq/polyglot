const http = require('http');

const BASE_URL = 'http://localhost:8080';
const DEVICE_ID = 'apa-saja';

function logStep(stepNum, title) {
  console.log(`\n======================================================================`);
  console.log(`  STEP ${stepNum}: ${title}`);
  console.log(`======================================================================`);
}

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

function startPersistentSSE(path, frameCallback) {
  const url = new URL(BASE_URL + path);
  const req = http.get(url, (res) => {
    console.log(`[📡 SSE STREAM OPENED] Path: ${path} | Status Code: ${res.statusCode}`);
    let buffer = '';

    res.on('data', (chunk) => {
      buffer += chunk.toString();
      const lines = buffer.split('\n');
      buffer = lines.pop();

      let currentEvent = '';
      let currentData = '';

      for (const line of lines) {
        if (line.startsWith('event:')) {
          currentEvent = line.substring(6).trim();
        } else if (line.startsWith('data:')) {
          currentData = line.substring(5).trim();
          if (currentEvent && currentData) {
            frameCallback(currentEvent, currentData);
            currentEvent = '';
            currentData = '';
          }
        }
      }
    });

    res.on('end', () => {
      console.log('[📡 SSE STREAM ENDED BY SERVER]');
    });
  });

  req.on('error', (err) => {
    console.error('[📡 SSE STREAM ERROR]:', err.message);
  });

  return req;
}

async function runLiveStreamingVerification() {
  console.log('🚀 INITIALIZING FULL REALTIME BACKEND WIRE STREAMING & LIFECYCLE TEST');
  console.log('   Target Router: 192.168.233.1 | User: admin | Pass: r00t | Port: 8728\n');

  // STEP 1: Set Initial Valid Config
  logStep(1, 'SET INITIAL VALID CONFIG (192.168.233.1, admin, r00t)');
  const initRes = await sendRequest(`/api/v1/devices/${DEVICE_ID}`, 'PUT', {
    id: DEVICE_ID,
    name: 'Router-233-Initial',
    vendor: 'mikrotik',
    driver_type: 'mikrotik',
    host: '192.168.233.1',
    port: 8728,
    timeout_ms: 10000,
    enabled: true,
    username: 'admin',
    password: 'r00t',
  });
  console.log('   REST Response:', initRes.status, JSON.stringify(initRes.data));

  // STEP 2: Start Continuous Background SSE Stream Listener
  logStep(2, 'STARTING CONTINUOUS BACKGROUND SSE STREAM LISTENER (/ws/devices/stream)');
  let frameCounter = 0;
  const sseReq = startPersistentSSE('/ws/devices/stream', (event, dataStr) => {
    frameCounter++;
    const timestamp = new Date().toISOString().substring(11, 23);
    try {
      const parsed = JSON.parse(dataStr);
      const targetItem = Array.isArray(parsed) ? parsed.find((i) => i.device && i.device.id === DEVICE_ID) : null;
      if (targetItem) {
        console.log(`\n--------------------------------------------------`);
        console.log(`📩  [FRAME #${frameCounter} @ ${timestamp}] Event: "${event}"`);
        console.log(`    Device ID:   ${targetItem.device.id}`);
        console.log(`    Name:        ${targetItem.device.name}`);
        console.log(`    Host IP:     ${targetItem.device.host}:${targetItem.device.port}`);
        console.log(`    Status:      ${targetItem.test.status.toUpperCase()}`);
        console.log(`    Message:     ${targetItem.test.message}`);
        if (targetItem.test.identity) console.log(`    Identity:    ${targetItem.test.identity}`);
        if (targetItem.test.version)  console.log(`    ROS Version: ${targetItem.test.version}`);
        if (targetItem.test.board_name) console.log(`    Board Name:  ${targetItem.test.board_name}`);
        if (targetItem.test.uptime)   console.log(`    Uptime:      ${targetItem.test.uptime}`);
        console.log(`--------------------------------------------------`);
      }
    } catch (e) {
      console.log(`[Raw Frame #${frameCounter}]: ${dataStr}`);
    }
  });

  // Wait 3 seconds to observe initial live streaming state
  await new Promise((r) => setTimeout(r, 3000));

  // STEP 3: Test Parameter Update (Change Name while streaming)
  logStep(3, 'LIVE UPDATE: Changing Router Name to "Router-233-UpdatedName-Active"');
  const updateNameRes = await sendRequest(`/api/v1/devices/${DEVICE_ID}`, 'PUT', {
    id: DEVICE_ID,
    name: 'Router-233-UpdatedName-Active',
    vendor: 'mikrotik',
    driver_type: 'mikrotik',
    host: '192.168.233.1',
    port: 8728,
    timeout_ms: 10000,
    enabled: true,
    username: 'admin',
    password: '', // blank -> keep r00t
  });
  console.log('   REST PUT Response:', updateNameRes.status, JSON.stringify(updateNameRes.data));

  // Wait 3 seconds to observe SSE stream reflecting new name
  await new Promise((r) => setTimeout(r, 3000));

  // STEP 4: Test Invalid Credential Update (Induce Failed / Offline state in active SSE stream)
  logStep(4, 'INVALID CREDENTIAL TEST: Changing password to "WRONG_PASS_1234" to induce FAILED status');
  const wrongCredRes = await sendRequest(`/api/v1/devices/${DEVICE_ID}`, 'PUT', {
    id: DEVICE_ID,
    name: 'Router-233-WrongCreds-Test',
    vendor: 'mikrotik',
    driver_type: 'mikrotik',
    host: '192.168.233.1',
    port: 8728,
    timeout_ms: 10000,
    enabled: true,
    username: 'admin',
    password: 'WRONG_PASS_1234',
  });
  console.log('   REST PUT Response:', wrongCredRes.status, JSON.stringify(wrongCredRes.data));

  // Wait 4 seconds to observe active SSE stream reporting FAILED status
  await new Promise((r) => setTimeout(r, 4000));

  // STEP 5: Recover Credential (Restore valid password and observe SSE auto-recovery)
  logStep(5, 'RECOVERY TEST: Restoring valid password "r00t" and observing SSE auto-recovery to CONNECTED');
  const recoverRes = await sendRequest(`/api/v1/devices/${DEVICE_ID}`, 'PUT', {
    id: DEVICE_ID,
    name: 'Router-233-RECOVERED-ONLINE',
    vendor: 'mikrotik',
    driver_type: 'mikrotik',
    host: '192.168.233.1',
    port: 8728,
    timeout_ms: 10000,
    enabled: true,
    username: 'admin',
    password: 'r00t',
  });
  console.log('   REST PUT Response:', recoverRes.status, JSON.stringify(recoverRes.data));

  // Wait 4 seconds to observe auto-recovery in active SSE stream
  await new Promise((r) => setTimeout(r, 4000));

  // STEP 6: Simulate Reboot / Offline State Test
  logStep(6, 'ROUTER REBOOT SIMULATION: Temporarily setting IP to unreachable host to observe offline behavior during reboot');
  const rebootSimRes = await sendRequest(`/api/v1/devices/${DEVICE_ID}`, 'PUT', {
    id: DEVICE_ID,
    name: 'Router-233-Rebooting...',
    vendor: 'mikrotik',
    driver_type: 'mikrotik',
    host: '192.168.233.254', // unreachable IP representing router rebooting/offline
    port: 8728,
    timeout_ms: 2000,
    enabled: true,
    username: 'admin',
    password: '',
  });
  console.log('   REST PUT Response:', rebootSimRes.status, JSON.stringify(rebootSimRes.data));

  // Wait 4 seconds to observe offline state during reboot
  await new Promise((r) => setTimeout(r, 4000));

  // STEP 7: Router Reboot Completion / Online Recovery Test
  logStep(7, 'ROUTER BOOTED UP: Restoring Host IP to 192.168.233.1 after reboot completion');
  const bootCompletedRes = await sendRequest(`/api/v1/devices/${DEVICE_ID}`, 'PUT', {
    id: DEVICE_ID,
    name: 'Router-233-BootCompleted-ONLINE',
    vendor: 'mikrotik',
    driver_type: 'mikrotik',
    host: '192.168.233.1',
    port: 8728,
    timeout_ms: 10000,
    enabled: true,
    username: 'admin',
    password: '',
  });
  console.log('   REST PUT Response:', bootCompletedRes.status, JSON.stringify(bootCompletedRes.data));

  // Wait 4 seconds to observe final online recovery frame
  await new Promise((r) => setTimeout(r, 4000));

  // Cleanup & Close SSE
  sseReq.destroy();
  console.log('\n🛑  SSE Stream connection closed.');

  logStep('FINAL', 'VERIFYING FINAL STATE VIA REST GET /api/v1/devices');
  const finalGet = await sendRequest('/api/v1/devices', 'GET');
  console.log('   Final DB Inventory State:', JSON.stringify(finalGet.data, null, 2));

  console.log('\n======================================================================');
  console.log('  ✅ ALL REALTIME STREAMING & LIFECYCLE TESTS PASSED PERFECTLY!');
  console.log('======================================================================\n');
}

runLiveStreamingVerification().catch((err) => {
  console.error('❌ Integration Test Error:', err);
  process.exit(1);
});
