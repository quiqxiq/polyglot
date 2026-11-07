'use strict';

const test = require('node:test');
const assert = require('node:assert/strict');
const DialMenu = require('./menu.js');

// Backend returns flat permissions as "resource:action" where the action is
// always the policy action ("*" for every seeded role). Normalize test input
// such as "device:read" to "device:read:*" to mirror GetImplicitPermissionsForUser.
function normalize(granted) {
  return granted.endsWith(':*') ? granted : `${granted}:*`;
}

function canFrom(...permissions) {
  const flat = permissions.map(normalize);
  return (required) => DialMenu.hasPermission(flat, required);
}

function visibleIds(...permissions) {
  const ids = [];
  const walk = (nodes) => {
    (nodes || []).forEach((node) => {
      ids.push(node.id);
      walk(node.children);
    });
  };
  walk(DialMenu.visibleTree(canFrom(...permissions)));
  return ids;
}

test('parent denial hides every descendant in hierarchy', () => {
  const customTree = [
    {
      id: 'restricted-group',
      type: 'group',
      allOf: ['custom:parent:permission'],
      children: [
        { id: 'child-page', type: 'page', href: 'child.html', allOf: ['custom:child:permission'] }
      ]
    }
  ];
  const can = canFrom('custom:child:permission');
  const resolved = DialMenu.resolve(customTree, can);
  assert.deepEqual(resolved, []);
});

test('allowed parent keeps only permitted children', () => {
  const can = canFrom('dial:ppp-secrets:read');

  assert.equal(DialMenu.canAccess('operations', can), true);
  assert.equal(DialMenu.canAccess('ppp', can), true);
  assert.equal(DialMenu.canAccess('ppp-secrets', can), true);
  assert.equal(DialMenu.canAccess('ppp-active', can), false);
  assert.equal(DialMenu.canAccess('ppp-inactive', can), false);
  assert.equal(DialMenu.canAccess('ppp-profiles', can), false);
  assert.equal(DialMenu.canAccess('monitoring', can), false);
});

test('group without visible children is removed', () => {
  const ids = visibleIds('custom:unrelated:permission');
  assert.deepEqual(ids, []);
});

test('allOf requires every permission', () => {
  const node = { id: 'multi', type: 'page', allOf: ['perm:a', 'perm:b'] };
  const missing = canFrom('perm:a');
  const complete = canFrom('perm:a', 'perm:b');

  assert.equal(DialMenu.checkNode(node, missing), false);
  assert.equal(DialMenu.checkNode(node, complete), true);
});

test('anyOf requires at least one permission', () => {
  const node = { id: 'choice', type: 'page', anyOf: ['perm:x', 'perm:y'] };

  assert.equal(DialMenu.checkNode(node, canFrom('perm:z')), false);
  assert.equal(DialMenu.checkNode(node, canFrom('perm:x')), true);
  assert.equal(DialMenu.checkNode(node, canFrom('perm:y')), true);
});

test('empty anyOf means no additional requirement', () => {
  const node = { id: 'x', type: 'page', anyOf: [] };

  assert.equal(DialMenu.checkNode(node, () => false), true);
});

test('menuPermission must also pass', () => {
  const node = { id: 'x', type: 'page', menuPermission: 'dial:custom:read', allOf: [] };

  assert.equal(DialMenu.checkNode(node, () => false), false);
  assert.equal(DialMenu.checkNode(node, (p) => p === 'dial:custom:read'), true);
});

test('firstAccessibleRoute skips forbidden pages', () => {
  assert.equal(DialMenu.firstAccessibleRoute(canFrom('dial:settings:read')), 'settings.html');
  assert.equal(DialMenu.firstAccessibleRoute(canFrom('dial:dashboard:read')), 'index.html');
  assert.equal(DialMenu.firstAccessibleRoute(canFrom('dial:ppp-active:read')), 'ppp-active.html');
  assert.equal(DialMenu.firstAccessibleRoute(canFrom('customer:read')), null);
});

test('canAction delegates directly to page access', () => {
  const denied = canFrom('dial:logs:read');
  const pageGranted = canFrom('dial:ppp-active:read');

  assert.equal(DialMenu.canAction('ppp-active', 'kick', denied), false);
  assert.equal(DialMenu.canAction('ppp-active', 'kick', pageGranted), true);
  assert.equal(DialMenu.canAction('ppp-active', 'ping', pageGranted), true);
});

test('settings is unified under dial:settings:read', () => {
  const readOnly = canFrom('dial:settings:read');
  const denied = canFrom('dial:ppp-active:read');

  assert.equal(DialMenu.canAccess('settings', readOnly), true);
  assert.equal(DialMenu.canAccess('settings-users', readOnly), true);
  assert.equal(DialMenu.canAccess('settings-devices', readOnly), true);
  assert.equal(DialMenu.canAction('settings', 'manageUsers', readOnly), true);
  assert.equal(DialMenu.canAction('settings-devices', 'assign', readOnly), true);

  assert.equal(DialMenu.canAccess('settings', denied), false);
  assert.equal(DialMenu.canAccess('settings-users', denied), false);
  assert.equal(DialMenu.canAccess('settings-devices', denied), false);
});

test('wildcard resource matches permissions', () => {
  assert.equal(DialMenu.matchesPermission('dial:device:.*:*', 'dial:device:read'), true);
  assert.equal(DialMenu.matchesPermission('dial:.*:*', 'dial:ppp-active:read'), true);
  assert.equal(DialMenu.matchesPermission('.*:*', 'dial:ppp-active:read'), true);
});

test('profile policy does not leak into ppp:profile permission', () => {
  assert.equal(DialMenu.matchesPermission('profile:.*:*', 'profile:read'), true);
  assert.equal(DialMenu.matchesPermission('profile:.*:*', 'profile:write'), true);
  assert.equal(DialMenu.matchesPermission('profile:.*:*', 'dial:ppp-profiles:read'), false);
});

test('owner wildcard sees full menu and actions', () => {
  const owner = canFrom('.*:*');
  const ids = visibleIds('.*:*');

  assert.equal(DialMenu.hasAnyMenu(owner), true);
  assert.equal(ids.includes('dashboard'), true);
  assert.equal(ids.includes('settings'), true);
  assert.equal(DialMenu.canAccess('monitor', owner), true);
  assert.equal(DialMenu.canAction('ppp-active', 'kick', owner), true);
  assert.equal(DialMenu.firstAccessibleRoute(owner), 'index.html');
});

test('agent-like permissions produce no menu', () => {
  const can = canFrom(
    'ppp:secret:read:*',
    'profile:.*:*',
    'customer:read:*',
    'billing:read:*'
  );

  assert.equal(DialMenu.hasAnyMenu(can), false);
  assert.equal(DialMenu.firstAccessibleRoute(can), null);
});

test('teknisi-like permissions show 7 operational pages and hide settings', () => {
  const can = canFrom(
    'dial:dashboard:read:*',
    'dial:ppp-active:read:*',
    'dial:ppp-non-active:read:*',
    'dial:ppp-secrets:read:*',
    'dial:ppp-profiles:read:*',
    'dial:logs:read:*',
    'dial:monitor:read:*'
  );

  assert.equal(DialMenu.canAccess('dashboard', can), true);
  assert.equal(DialMenu.canAccess('ppp-active', can), true);
  assert.equal(DialMenu.canAccess('settings', can), false);
  assert.equal(DialMenu.canAction('ppp-active', 'kick', can), true);
  assert.equal(DialMenu.canAction('ppp-active', 'ping', can), true);
  assert.equal(DialMenu.canAction('ppp-inactive', 'toggleSecret', can), true);
  assert.equal(DialMenu.canAction('ppp-secrets', 'toggleSecret', can), true);
});

test('bottom nav only contains visible leaf pages', () => {
  const items = DialMenu.bottomNavItems(canFrom('dial:dashboard:read', 'dial:ppp-active:read'));

  assert.deepEqual(items.map((item) => item.id), ['dashboard', 'ppp-active']);
});

test('findPageByHref normalizes hash and query', () => {
  assert.equal(DialMenu.findPageByHref('settings.html#users').id, 'settings');
  assert.equal(DialMenu.findPageByHref('ppp-active.html?search=budi').id, 'ppp-active');
});

test('equivalent core permissions unlock web dial pages', () => {
  const can = canFrom('device:read', 'ppp:profile:read', 'log:read');

  assert.equal(DialMenu.canAccess('ppp-profiles', can), true);
  assert.equal(DialMenu.canAccess('logs', can), true);
  assert.equal(DialMenu.canAccess('ppp-secrets', can), false);
  assert.equal(DialMenu.canAccess('monitor', can), false);
});

test('revoking permissions hides web dial menus cleanly', () => {
  // Only has dashboard and ppp-active, ppp-profiles/logs/monitor revoked
  const revoked = canFrom('dial:dashboard:read', 'dial:ppp-active:read');

  assert.equal(DialMenu.canAccess('ppp-profiles', revoked), false);
  assert.equal(DialMenu.canAccess('ppp-secrets', revoked), false);
  assert.equal(DialMenu.canAccess('logs', revoked), false);
  assert.equal(DialMenu.canAccess('monitor', revoked), false);

  const ids = visibleIds('dial:dashboard:read', 'dial:ppp-active:read');
  assert.equal(ids.includes('ppp-profiles'), false);
  assert.equal(ids.includes('ppp-secrets'), false);
  assert.equal(ids.includes('logs'), false);
  assert.equal(ids.includes('monitor'), false);
  assert.equal(ids.includes('ppp-active'), true);
  assert.equal(ids.includes('dashboard'), true);
});

test('each of the 8 pages is unlocked 1:1 by its dedicated permission', () => {
  const pages = [
    { id: 'dashboard', perm: 'dial:dashboard:read' },
    { id: 'ppp-active', perm: 'dial:ppp-active:read' },
    { id: 'ppp-inactive', perm: 'dial:ppp-non-active:read' },
    { id: 'ppp-secrets', perm: 'dial:ppp-secrets:read' },
    { id: 'ppp-profiles', perm: 'dial:ppp-profiles:read' },
    { id: 'logs', perm: 'dial:logs:read' },
    { id: 'monitor', perm: 'dial:monitor:read' },
    { id: 'settings', perm: 'dial:settings:read' },
  ];

  for (const p of pages) {
    const can = canFrom(p.perm);
    assert.equal(DialMenu.canAccess(p.id, can), true, `Page ${p.id} should be accessible with ${p.perm}`);
    for (const other of pages) {
      if (other.id !== p.id) {
        assert.equal(DialMenu.canAccess(other.id, can), false, `Page ${other.id} should NOT be accessible with ${p.perm}`);
      }
    }
  }
});
