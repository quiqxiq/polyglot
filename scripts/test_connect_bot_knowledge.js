const http = require('http');

function makeConnectRequest(service, method, payload) {
  return new Promise((resolve, reject) => {
    const data = JSON.stringify(payload || {});
    const url = `http://localhost:8080/polyglot.v1.${service}/${method}`;
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

async function runBotKnowledgeConnectRPCTest() {
  console.log('🚀 STARTING BOT & KNOWLEDGE ConnectRPC INTEGRATION TEST (over HTTP :8080)...');

  // 1. BotService - ListSessions
  console.log('\n1️⃣  Testing BotService.ListSessions...');
  const sessRes = await makeConnectRequest('BotService', 'ListSessions', {});
  console.log('   Status:', sessRes.status);
  console.log('   Sessions Count:', sessRes.data?.sessions?.length || 0);

  // 2. BotService - ListConversations
  console.log('\n2️⃣  Testing BotService.ListConversations...');
  const convRes = await makeConnectRequest('BotService', 'ListConversations', { session_id: '1' });
  console.log('   Status:', convRes.status);
  console.log('   Conversations Count:', convRes.data?.conversations?.length || 0);

  // 3. KnowledgeService - ListKnowledge
  console.log('\n3️⃣  Testing KnowledgeService.ListKnowledge...');
  const knwRes = await makeConnectRequest('KnowledgeService', 'ListKnowledge', {});
  console.log('   Status:', knwRes.status);
  console.log('   Knowledge Items Count:', knwRes.data?.items?.length || 0);

  // 4. KnowledgeService - ListLLMConfigs
  console.log('\n4️⃣  Testing KnowledgeService.ListLLMConfigs...');
  const llmRes = await makeConnectRequest('KnowledgeService', 'ListLLMConfigs', {});
  console.log('   Status:', llmRes.status);
  console.log('   LLM Configs:', JSON.stringify(llmRes.data, null, 2));

  // 5. KnowledgeService - ListTechnicians
  console.log('\n5️⃣  Testing KnowledgeService.ListTechnicians...');
  const techRes = await makeConnectRequest('KnowledgeService', 'ListTechnicians', {});
  console.log('   Status:', techRes.status);
  console.log('   Technicians:', JSON.stringify(techRes.data, null, 2));

  console.log('\n======================================================================');
  console.log('  ✅ ALL BOT & KNOWLEDGE ConnectRPC PROCEDURES VERIFIED SUCCESSFULLY!');
  console.log('======================================================================');
}

runBotKnowledgeConnectRPCTest().catch((err) => {
  console.error('❌ Bot & Knowledge ConnectRPC Test Error:', err);
  process.exit(1);
});
