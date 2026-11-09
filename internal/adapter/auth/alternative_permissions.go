package auth

// AlternativeProcedurePermissions maps procedures that can also be authorized
// by alternative permissions (e.g. portal-specific permissions like dial:*).
var AlternativeProcedurePermissions = map[string][]string{
	// Device & Monitor
	"/polyglot.v1.DeviceService/ListDevices": {
		"dial:device:read",
		"dial:dashboard:read",
		"dial:ppp-active:read",
		"dial:ppp-non-active:read",
		"dial:ppp-secrets:read",
		"dial:ppp-profiles:read",
		"dial:logs:read",
		"dial:monitor:read",
		"dial:settings:read",
	},
	"/polyglot.v1.DeviceService/GetDevice": {
		"dial:device:read",
		"dial:dashboard:read",
		"dial:ppp-active:read",
		"dial:ppp-non-active:read",
		"dial:ppp-secrets:read",
		"dial:ppp-profiles:read",
		"dial:logs:read",
		"dial:monitor:read",
		"dial:settings:read",
	},
	"/polyglot.v1.DeviceService/TestDeviceConnection": {
		"dial:monitor:read",
		"dial:dashboard:read",
		"dial:monitor:diagnostic",
		"dial:dashboard:ping",
	},
	"/polyglot.v1.DeviceService/StreamDeviceStatus": {
		"dial:device:read",
		"dial:dashboard:read",
		"dial:ppp-active:read",
		"dial:ppp-non-active:read",
		"dial:ppp-secrets:read",
		"dial:ppp-profiles:read",
		"dial:logs:read",
		"dial:monitor:read",
		"dial:settings:read",
	},
	"/polyglot.v1.DeviceService/StreamPing": {
		"dial:dashboard:read",
		"dial:ppp-active:read",
		"dial:monitor:read",
		"dial:dashboard:ping",
		"dial:ppp-active:ping",
	},
	"/polyglot.v1.DeviceService/StreamInterfaceTraffic": {
		"dial:monitor:read",
		"dial:dashboard:read",
	},

	// NetworkMonitorService
	"/polyglot.v1.NetworkMonitorService/StreamTraffic": {
		"dial:monitor:read",
		"dial:dashboard:read",
	},
	"/polyglot.v1.NetworkMonitorService/StreamResource": {
		"dial:monitor:read",
		"dial:dashboard:read",
	},
	"/polyglot.v1.NetworkMonitorService/StreamSystemSnapshot": {
		"dial:monitor:read",
		"dial:dashboard:read",
	},
	"/polyglot.v1.NetworkMonitorService/StreamInterfaceEthernet": {
		"dial:monitor:read",
	},
	"/polyglot.v1.NetworkMonitorService/StreamQueueStats": {
		"dial:monitor:read",
	},
	"/polyglot.v1.NetworkMonitorService/StreamLogs": {
		"dial:logs:read",
	},
	"/polyglot.v1.NetworkMonitorService/StreamPPPActive": {
		"dial:ppp-active:read",
		"dial:dashboard:read",
	},
	"/polyglot.v1.NetworkMonitorService/StreamPPPInactive": {
		"dial:ppp-non-active:read",
		"dial:dashboard:read",
	},

	// PPPService
	"/polyglot.v1.PPPService/ListActiveSessions": {
		"dial:ppp-active:read",
		"dial:dashboard:read",
	},
	"/polyglot.v1.PPPService/StreamActiveSessions": {
		"dial:ppp-active:read",
		"dial:dashboard:read",
	},
	"/polyglot.v1.PPPService/KickActiveSession": {
		"dial:ppp-active:read",
		"dial:dashboard:read",
		"dial:ppp-active:kick",
	},
	"/polyglot.v1.PPPService/KickActiveSessions": {
		"dial:ppp-active:read",
		"dial:dashboard:read",
		"dial:ppp-active:kick",
	},
	"/polyglot.v1.PPPService/ListInactiveSecrets": {
		"dial:ppp-non-active:read",
		"dial:dashboard:read",
	},
	"/polyglot.v1.PPPService/StreamInactiveSecrets": {
		"dial:ppp-non-active:read",
		"dial:dashboard:read",
	},
	"/polyglot.v1.PPPService/ListSecrets": {
		"dial:ppp-secrets:read",
		"dial:dashboard:read",
	},
	"/polyglot.v1.PPPService/GetSecret": {
		"dial:ppp-secrets:read",
		"dial:dashboard:read",
	},
	"/polyglot.v1.PPPService/CreateSecret": {
		"dial:ppp-secrets:read",
		"dial:ppp-secrets:write",
	},
	"/polyglot.v1.PPPService/UpdateSecret": {
		"dial:ppp-secrets:read",
		"dial:ppp-secrets:write",
	},
	"/polyglot.v1.PPPService/DeleteSecret": {
		"dial:ppp-secrets:read",
		"dial:ppp-secrets:write",
	},
	"/polyglot.v1.PPPService/SetSecretDisabled": {
		"dial:ppp-non-active:read",
		"dial:ppp-secrets:read",
		"dial:dashboard:read",
		"dial:ppp-non-active:toggle",
		"dial:ppp-secrets:write",
	},
	"/polyglot.v1.PPPService/ListProfiles": {
		"dial:ppp-profiles:read",
		"dial:ppp-secrets:read",
		"dial:dashboard:read",
	},
	"/polyglot.v1.PPPService/GetProfile": {
		"dial:ppp-profiles:read",
		"dial:dashboard:read",
	},

	// LogService
	"/polyglot.v1.LogService/StreamLogs": {
		"dial:logs:read",
	},
	"/polyglot.v1.LogService/ListLogs": {
		"dial:logs:read",
	},

	// UserService
	"/polyglot.v1.UserService/ListUsers": {
		"dial:settings:read",
		"dial:settings:users:read",
	},
	"/polyglot.v1.UserService/CreateUser": {
		"dial:settings:read",
		"dial:settings:users:manage",
	},
	"/polyglot.v1.UserService/UpdateUser": {
		"dial:settings:read",
		"dial:settings:users:manage",
	},
	"/polyglot.v1.UserService/ResetPassword": {
		"dial:settings:read",
		"dial:settings:users:manage",
	},
	"/polyglot.v1.UserService/ToggleActive": {
		"dial:settings:read",
		"dial:settings:users:manage",
	},
	"/polyglot.v1.UserService/DeleteUser": {
		"dial:settings:read",
		"dial:settings:users:manage",
	},
	"/polyglot.v1.UserService/AssignDevicesToUser": {
		"dial:settings:read",
		"dial:settings:devices:manage",
	},
	"/polyglot.v1.UserService/ListUserAccessibleDevices": {
		"dial:settings:read",
		"dial:settings:devices:read",
	},
}
