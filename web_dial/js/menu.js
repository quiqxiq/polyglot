/**
 * Polyglot Web Dial Menu Registry & Permission Resolver
 *
 * Satu-satunya sumber kebenaran untuk hierarki menu dan permission yang
 * dibutuhkan setiap halaman/aksi. Resolver bersifat murni (menerima predicate
 * `can`) agar bisa diuji tanpa browser maupun localStorage.
 *
 * Cascade: jika sebuah node ditolak, seluruh descendant-nya otomatis ditolak.
 */

(function (root, factory) {
  const DialMenu = factory();
  if (typeof module !== 'undefined' && module.exports) {
    module.exports = DialMenu;
  }
  if (root) {
    root.DialMenu = DialMenu;
  }
})(typeof globalThis !== 'undefined' ? globalThis : this, function () {
  'use strict';

  const ICONS = {
    dashboard: '<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M15 21v-8a1 1 0 0 0-1-1h-4a1 1 0 0 0-1 1v8"/><path d="M3 10a2 2 0 0 1 .709-1.528l7-6a2 2 0 0 1 2.582 0l7 6A2 2 0 0 1 21 10v9a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/></svg>',
    wifi: '<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 20h.01"/><path d="M2 8.82a15 15 0 0 1 20 0"/><path d="M5 12.859a10 10 0 0 1 14 0"/><path d="M8.5 16.429a5 5 0 0 1 7 0"/></svg>',
    'wifi-off': '<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 20h.01"/><path d="M8.5 16.429a5 5 0 0 1 7 0"/><path d="M5 12.859a10 10 0 0 1 5.17-2.69"/><path d="M19 12.859a10 10 0 0 0-2.007-1.523"/><path d="M2 8.82a15 15 0 0 1 4.177-2.643"/><path d="M22 8.82a15 15 0 0 0-11.288-3.764"/><path d="m2 2 20 20"/></svg>',
    scroll: '<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M15 12h-5"/><path d="M15 8h-5"/><path d="M19 17V5a2 2 0 0 0-2-2H4"/><path d="M8 21h12a2 2 0 0 0 2-2v-1a1 1 0 0 0-1-1H11a1 1 0 0 0-1 1v2a1 1 0 0 0 1 1z"/></svg>',
    monitor: '<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="16" y="16" width="6" height="6" rx="1"/><rect x="2" y="16" width="6" height="6" rx="1"/><rect x="9" y="2" width="6" height="6" rx="1"/><path d="M5 16v-3a1 1 0 0 1 1-1h12a1 1 0 0 1 1 1v3"/><path d="M12 12V8"/></svg>',
    key: '<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M2 18v3c0 .6.4 1 1 1h4v-3h3v-3h2l1.4-1.4a6.5 6.5 0 1 0-4-4Z"/><circle cx="16.5" cy="7.5" r=".5" fill="currentColor"/></svg>',
    layers: '<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m12.83 2.18a2 2 0 0 0-1.66 0L2.6 6.08a1 1 0 0 0 0 1.83l8.58 3.91a2 2 0 0 0 1.66 0l8.58-3.9a1 1 0 0 0 0-1.83Z"/><path d="m22 17.65-9.17 4.16a2 2 0 0 1-1.66 0L2 17.65"/><path d="m22 12.65-9.17 4.16a2 2 0 0 1-1.66 0L2 12.65"/></svg>',
    settings: '<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z"/><circle cx="12" cy="12" r="3"/></svg>',
    activity: '<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/></svg>',
    users: '<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M22 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/></svg>',
    router: '<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="20" height="8" x="2" y="14" rx="2"/><path d="M6.01 18h.01"/><path d="M10.01 18h.01"/><path d="M15 10v4"/><path d="M17.84 7.17a4 4 0 0 0-5.66 0"/><path d="M20.66 4.34a8 8 0 0 0-11.31 0"/></svg>',
    default: '<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/></svg>'
  };

  const TREE = [
    {
      id: 'dashboard',
      type: 'page',
      labelKey: 'nav_dashboard',
      icon: 'dashboard',
      href: 'index.html',
      bottomNav: true,
      allOf: ['dial:dashboard:read', 'dial:device:read'],
      anyOf: [
        'dial:monitor:read',
        'dial:ppp-active:read',
        'dial:ppp-non-active:read',
        'dial:ppp-secrets:read',
        'dial:ppp-profiles:read'
      ],
      actions: {
        snapshot: 'dial:monitor:read',
        search: 'dial:ppp-active:read',
        kick: 'dial:ppp-active:kick',
        toggleSecret: 'dial:ppp-non-active:toggle',
        ping: 'dial:dashboard:ping'
      }
    },
    {
      id: 'operations',
      type: 'group',
      labelKey: 'group_operations',
      icon: 'activity',
      allOf: ['dial:device:read'],
      anyOf: [
        'dial:ppp-active:read',
        'dial:ppp-non-active:read',
        'dial:ppp-secrets:read',
        'dial:ppp-profiles:read',
        'dial:logs:read',
        'dial:monitor:read'
      ],
      children: [
        {
          id: 'ppp',
          type: 'group',
          labelKey: 'group_ppp',
          icon: 'wifi',
          anyOf: [
            'dial:ppp-active:read',
            'dial:ppp-non-active:read',
            'dial:ppp-secrets:read',
            'dial:ppp-profiles:read'
          ],
          children: [
            {
              id: 'ppp-active',
              type: 'page',
              labelKey: 'nav_ppp_active',
              icon: 'wifi',
              href: 'ppp-active.html',
              bottomNav: true,
              allOf: ['dial:ppp-active:read'],
              actions: {
                kick: 'dial:ppp-active:kick',
                ping: 'dial:ppp-active:ping'
              }
            },
            {
              id: 'ppp-inactive',
              type: 'page',
              labelKey: 'nav_ppp_inactive',
              icon: 'wifi-off',
              href: 'ppp-non-active.html',
              bottomNav: true,
              allOf: ['dial:ppp-non-active:read'],
              actions: {
                toggleSecret: 'dial:ppp-non-active:toggle'
              }
            },
            {
              id: 'ppp-secrets',
              type: 'page',
              labelKey: 'nav_ppp_secrets',
              icon: 'key',
              href: 'ppp-secrets.html',
              allOf: ['dial:ppp-secrets:read'],
              actions: {
                toggleSecret: 'dial:ppp-secrets:write',
                openActive: 'dial:ppp-active:read'
              }
            },
            {
              id: 'ppp-profiles',
              type: 'page',
              labelKey: 'nav_ppp_profiles',
              icon: 'layers',
              href: 'ppp-profiles.html',
              allOf: ['dial:ppp-profiles:read']
            }
          ]
        },
        {
          id: 'monitoring',
          type: 'group',
          labelKey: 'group_monitoring',
          icon: 'monitor',
          anyOf: ['dial:logs:read', 'dial:monitor:read'],
          children: [
            {
              id: 'logs',
              type: 'page',
              labelKey: 'nav_logs',
              icon: 'scroll',
              href: 'mikrotik-logs.html',
              bottomNav: true,
              allOf: ['dial:logs:read']
            },
            {
              id: 'monitor',
              type: 'page',
              labelKey: 'nav_monitor',
              icon: 'monitor',
              href: 'monitor-interface.html',
              bottomNav: true,
              allOf: ['dial:monitor:read', 'dial:monitor:diagnostic'],
              actions: {
                testConnection: 'dial:monitor:diagnostic'
              }
            }
          ]
        }
      ]
    },
    {
      id: 'system',
      type: 'group',
      labelKey: 'group_system',
      icon: 'settings',
      anyOf: ['dial:settings:users:read', 'dial:settings:devices:read'],
      children: [
        {
          id: 'settings',
          type: 'page',
          labelKey: 'nav_settings',
          icon: 'settings',
          href: 'settings.html',
          anyOf: ['dial:settings:users:read', 'dial:settings:devices:read'],
          actions: {
            manageUsers: 'dial:settings:users:manage',
            listDevices: 'dial:settings:devices:read',
            assignDevices: ['dial:settings:users:manage', 'dial:settings:devices:read']
          },
          children: [
            {
              id: 'settings-users',
              type: 'tab',
              labelKey: 'nav_settings_users',
              icon: 'users',
              href: 'settings.html#users',
              allOf: ['dial:settings:users:read'],
              actions: {
                manage: 'dial:settings:users:manage'
              }
            },
            {
              id: 'settings-devices',
              type: 'tab',
              labelKey: 'nav_settings_devices',
              icon: 'router',
              href: 'settings.html#devices',
              allOf: ['dial:settings:devices:read', 'dial:settings:devices:manage'],
              actions: {
                assign: ['dial:settings:users:manage', 'dial:settings:devices:read'],
                listDevices: 'dial:settings:devices:read'
              }
            }
          ]
        }
      ]
    }
  ];

  function permissionList(value) {
    if (!value) return [];
    return Array.isArray(value) ? value : [value];
  }

  /**
   * Meniru enforcement Casbin: flat permission "resource:action" (bisa regex
   * pada resource) dicocokkan dengan permission yang diminta. Regex di-anchor
   * agar policy seperti "profile:.*" tidak bocor ke "ppp:profile:read".
   */
  function matchesPermission(granted, required) {
    if (typeof granted !== 'string' || typeof required !== 'string') return false;
    const reqIdx = required.lastIndexOf(':');
    if (reqIdx <= 0) return false;
    const reqAction = required.slice(reqIdx + 1);

    const grantIdx = granted.lastIndexOf(':');
    if (grantIdx <= 0) return false;
    const resource = granted.slice(0, grantIdx);
    const action = granted.slice(grantIdx + 1);

    if (action !== '*' && action !== reqAction) return false;

    try {
      return new RegExp('^(?:' + resource + ')$').test(required);
    } catch (_) {
      return resource === required;
    }
  }

  function hasPermission(permissions, required) {
    if (!Array.isArray(permissions) || permissions.length === 0) return false;
    return permissions.some((granted) => matchesPermission(granted, required));
  }

  function canAll(permissions, required) {
    return permissionList(required).every((p) => hasPermission(permissions, p));
  }

  function canAny(permissions, required) {
    const list = permissionList(required);
    if (list.length === 0) return true;
    return list.some((p) => hasPermission(permissions, p));
  }

  function checkNode(node, canFn) {
    if (!node) return false;
    if (node.menuPermission && !canFn(node.menuPermission)) return false;
    if (permissionList(node.allOf).some((p) => !canFn(p))) return false;
    if (permissionList(node.anyOf).length > 0 && !permissionList(node.anyOf).some((p) => canFn(p))) return false;
    return true;
  }

  function resolveNode(node, parentAllowed, canFn) {
    if (!parentAllowed || !checkNode(node, canFn)) return null;

    const children = (node.children || [])
      .map((child) => resolveNode(child, true, canFn))
      .filter(Boolean);

    if (node.type === 'group' && children.length === 0) return null;

    const resolved = Object.assign({}, node, { children });
    return resolved;
  }

  function resolve(tree, canFn) {
    const source = Array.isArray(tree) ? tree : TREE;
    return source.map((node) => resolveNode(node, true, canFn)).filter(Boolean);
  }

  function visibleTree(canFn) {
    return resolve(TREE, canFn);
  }

  function walk(nodes, visitor) {
    (nodes || []).forEach((node) => {
      visitor(node);
      if (node.children && node.children.length > 0) {
        walk(node.children, visitor);
      }
    });
  }

  function findNode(menuId, tree) {
    let found = null;
    walk(Array.isArray(tree) ? tree : TREE, (node) => {
      if (!found && node.id === menuId) found = node;
    });
    return found;
  }

  function findPath(menuId, tree) {
    const source = Array.isArray(tree) ? tree : TREE;
    const stack = [];

    function visit(nodes) {
      for (const node of nodes) {
        stack.push(node);
        if (node.id === menuId) return true;
        if (node.children && visit(node.children)) return true;
        stack.pop();
      }
      return false;
    }

    return visit(source) ? stack.slice() : null;
  }

  function canAccess(menuId, canFn) {
    const path = findPath(menuId);
    if (!path || path.length === 0) return false;
    return path.every((node) => checkNode(node, canFn));
  }

  function canAction(menuId, actionKey, canFn) {
    const node = findNode(menuId);
    if (!node || !node.actions || !node.actions[actionKey]) return false;
    if (!canAccess(menuId, canFn)) return false;
    return permissionList(node.actions[actionKey]).every((p) => canFn(p));
  }

  function flattenPages(tree) {
    const pages = [];
    walk(tree, (node) => {
      if (node.type === 'page' && node.href) pages.push(node);
    });
    return pages;
  }

  function firstAccessibleRoute(canFn) {
    const pages = flattenPages(visibleTree(canFn));
    return pages.length > 0 ? pages[0].href : null;
  }

  function bottomNavItems(canFn) {
    return flattenPages(visibleTree(canFn)).filter((node) => node.bottomNav);
  }

  function findPageByHref(href) {
    const normalized = String(href || '').split('#')[0].split('?')[0] || 'index.html';
    let found = null;
    walk(TREE, (node) => {
      if (found || node.type !== 'page' || !node.href) return;
      const nodeHref = node.href.split('#')[0].split('?')[0];
      if (nodeHref === normalized) found = node;
    });
    return found;
  }

  function hasAnyMenu(canFn) {
    return flattenPages(visibleTree(canFn)).length > 0;
  }

  function icon(name) {
    return ICONS[name] || ICONS.default;
  }

  return {
    ICONS,
    tree: TREE,
    resolve,
    visibleTree,
    checkNode,
    matchesPermission,
    hasPermission,
    canAll,
    canAny,
    findNode,
    findPath,
    findPageByHref,
    canAccess,
    canAction,
    firstAccessibleRoute,
    bottomNavItems,
    hasAnyMenu,
    flattenPages,
    icon
  };
});
