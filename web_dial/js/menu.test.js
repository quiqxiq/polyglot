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
  return (required) => flat.some((granted) => DialMenu.matchesPermission(granted, required));
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

test('parent denial hides every descendant even when child permission is granted', () => {
  const can = canFrom('dial:ppp-active:read');

  assert.equal(DialMenu.canAccess('ppp-active', can), false);
  assert.equal(DialMenu.canAccess('operations', can), false);
  assert.equal(DialMenu.canAccess('ppp', can), false);
  assert.deepEqual(visibleIds('dial:ppp-active:read'), []);
});

test('allowed parent keeps only permitted children', () => {
  const can = canFrom('dial:device:read', 'dial:ppp-secrets:read');

  assert.equal(DialMenu.canAccess('operations', can), true);
  assert.equal(DialMenu.canAccess('ppp', can), true);
  assert.equal(DialMenu.canAccess('ppp-secrets', can), true);
  assert.equal(DialMenu.canAccess('ppp-active', can), false);
  assert.equal(DialMenu.canAccess('ppp-inactive', can), false);
  assert.equal(DialMenu.canAccess('ppp-profiles', can), false);
  assert.equal(DialMenu.canAccess('monitoring', can), false);
});

test('group without visible children is removed', () => {
  const ids = visibleIds('dial:device:read');

  assert.equal(ids.includes('operations'), false);
  assert.equal(ids.includes('ppp'), false);
  assert.equal(ids.includes('monitoring'), false);
  assert.deepEqual(ids, []);
});

test('allOf requires every permission', () => {
  const missing = canFrom('dial:device:read', 'dial:monitor:read');
  const complete = canFrom('dial:device:read', 'dial:monitor:read', 'dial:monitor:diagnostic');

  assert.equal(DialMenu.canAccess('monitor', missing), false);
  assert.equal(DialMenu.canAccess('monitor', complete), true);
});

test('anyOf requires at least one permission', () => {
  const dashboard = DialMenu.findNode('dashboard');

  assert.equal(DialMenu.checkNode(dashboard, canFrom('dial:dashboard:read', 'dial:device:read')), false);
  assert.equal(DialMenu.checkNode(dashboard, canFrom('dial:dashboard:read', 'dial:device:read', 'dial:monitor:read')), true);
  assert.equal(DialMenu.checkNode(dashboard, canFrom('dial:dashboard:read', 'dial:device:read', 'dial:ppp-profiles:read')), true);
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
  assert.equal(DialMenu.firstAccessibleRoute(canFrom('dial:settings:users:read')), 'settings.html');
  assert.equal(DialMenu.firstAccessibleRoute(canFrom('dial:dashboard:read', 'dial:device:read', 'dial:ppp-secrets:read')), 'index.html');
  assert.equal(DialMenu.firstAccessibleRoute(canFrom('customer:read')), null);
});

test('canAction requires menu access and action permission', () => {
  const actionOnly = canFrom('dial:ppp-active:kick');
  const pageOnly = canFrom('dial:device:read', 'dial:ppp-active:read');
  const complete = canFrom('dial:device:read', 'dial:ppp-active:read', 'dial:ppp-active:kick');

  assert.equal(DialMenu.canAction('ppp-active', 'kick', actionOnly), false);
  assert.equal(DialMenu.canAction('ppp-active', 'kick', pageOnly), false);
  assert.equal(DialMenu.canAction('ppp-active', 'kick', complete), true);
});

test('array action permission requires all permissions', () => {
  const manageOnly = canFrom('dial:settings:users:read', 'dial:settings:users:manage');
  const complete = canFrom('dial:settings:users:read', 'dial:settings:users:manage', 'dial:settings:devices:read');

  assert.equal(DialMenu.canAction('settings', 'assignDevices', manageOnly), false);
  assert.equal(DialMenu.canAction('settings', 'assignDevices', complete), true);
});

test('wildcard resource matches device permissions', () => {
  assert.equal(DialMenu.matchesPermission('dial:device:.*:*', 'dial:device:read'), true);
  assert.equal(DialMenu.matchesPermission('dial:.*:*', 'dial:ppp-active:kick'), true);
  assert.equal(DialMenu.matchesPermission('.*:*', 'dial:ppp-active:kick'), true);
});

test('profile policy does not leak into ppp:profile permission', () => {
  assert.equal(DialMenu.matchesPermission('profile:.*:*', 'profile:read'), true);
  assert.equal(DialMenu.matchesPermission('profile:.*:*', 'profile:write'), true);
  assert.equal(DialMenu.matchesPermission('profile:.*:*', 'dial:ppp-profiles:read'), false);
});

test('settings tabs are gated independently', () => {
  const readOnly = canFrom('dial:settings:users:read');
  const routerManager = canFrom('dial:settings:users:read', 'dial:settings:users:manage', 'dial:settings:devices:read', 'dial:settings:devices:manage');

  assert.equal(DialMenu.canAccess('settings', readOnly), true);
  assert.equal(DialMenu.canAccess('settings-users', readOnly), true);
  assert.equal(DialMenu.canAccess('settings-devices', readOnly), false);
  assert.equal(DialMenu.canAction('settings', 'manageUsers', readOnly), false);
  assert.equal(DialMenu.canAction('settings', 'manageUsers', routerManager), true);
  assert.equal(DialMenu.canAccess('settings-devices', routerManager), true);
});

test('owner wildcard sees full menu and actions', () => {
  const owner = canFrom('.*:*');
  const ids = visibleIds('.*:*');

  assert.equal(DialMenu.hasAnyMenu(owner), true);
  assert.equal(ids.includes('settings'), true);
  assert.equal(ids.includes('settings-devices'), true);
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

test('teknisi-like permissions hide settings and kick actions', () => {
  const can = canFrom(
    'dial:dashboard:read:*',
    'dial:dashboard:ping:*',
    'dial:device:read:*',
    'dial:ppp-active:read:*',
    'dial:ppp-active:ping:*',
    'dial:ppp-non-active:read:*',
    'dial:ppp-secrets:read:*',
    'dial:ppp-profiles:read:*',
    'dial:logs:read:*',
    'dial:monitor:read:*',
    'dial:monitor:diagnostic:*'
  );

  assert.equal(DialMenu.canAccess('dashboard', can), true);
  assert.equal(DialMenu.canAccess('ppp-active', can), true);
  assert.equal(DialMenu.canAccess('settings', can), false);
  assert.equal(DialMenu.canAction('ppp-active', 'kick', can), false);
  assert.equal(DialMenu.canAction('ppp-active', 'ping', can), true);
  assert.equal(DialMenu.canAction('ppp-inactive', 'toggleSecret', can), false);
  assert.equal(DialMenu.canAction('ppp-secrets', 'toggleSecret', can), false);
});

test('bottom nav only contains visible leaf pages', () => {
  const items = DialMenu.bottomNavItems(canFrom('dial:dashboard:read', 'dial:device:read', 'dial:ppp-active:read'));

  assert.deepEqual(items.map((item) => item.id), ['dashboard', 'ppp-active']);
});

test('findPageByHref normalizes hash and query', () => {
  assert.equal(DialMenu.findPageByHref('settings.html#users').id, 'settings');
  assert.equal(DialMenu.findPageByHref('ppp-active.html?search=budi').id, 'ppp-active');
});
