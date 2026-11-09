// DEVIASI: code generator CLI tool for RouterOS schema inspection
package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// Inspector discovers the commands and argument schemas for a RouterOS path.
type Inspector interface {
	InspectMenu(ctx context.Context, rootMenu string) ([]*SubmenuSchema, error)
}

// LiveRESTInspector queries RouterOS v7 via its HTTP REST /console/inspect endpoint.
type LiveRESTInspector struct {
	BaseURL  string
	Username string
	Password string
	Client   *http.Client
}

// NewLiveRESTInspector creates an inspector targeting a live RouterOS REST endpoint.
func NewLiveRESTInspector(host string, port int, user, pass string, useTLS, insecure bool) *LiveRESTInspector {
	scheme := "http"
	if useTLS {
		scheme = "https"
	}
	baseURL := fmt.Sprintf("%s://%s:%d", scheme, host, port)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: insecure, //nolint:gosec // user option for testing against self-signed certs
		},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   15 * time.Second,
	}

	return &LiveRESTInspector{
		BaseURL:  baseURL,
		Username: user,
		Password: pass,
		Client:   client,
	}
}

func (l *LiveRESTInspector) postInspect(ctx context.Context, reqBody map[string]string) ([]map[string]any, error) {
	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/rest/console/inspect", l.BaseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if l.Username != "" || l.Password != "" {
		req.SetBasicAuth(l.Username, l.Password)
	}

	resp, err := l.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("POST /rest/console/inspect error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("RouterOS inspect error (status %d): %s", resp.StatusCode, string(b))
	}

	var results []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("decode inspect response: %w", err)
	}
	return results, nil
}

// InspectMenu walks the child submenus and their commands for rootMenu (e.g. "ppp").
func (l *LiveRESTInspector) InspectMenu(ctx context.Context, rootMenu string) ([]*SubmenuSchema, error) {
	normRoot := normalizePath(rootMenu)

	// Step 1: list child items under rootMenu
	children, err := l.postInspect(ctx, map[string]string{
		"request": "child",
		"path":    normRoot,
	})
	if err != nil {
		return nil, fmt.Errorf("inspect children of %s: %w", normRoot, err)
	}

	var subPaths []string
	var hasCommandsDirectly bool

	for _, c := range children {
		itemType, _ := c["type"].(string)
		if itemType == "self" {
			continue
		}
		childName, _ := c["name"].(string)
		childNodeType, _ := c["node-type"].(string)

		if childNodeType == "dir" || childNodeType == "path" {
			subPaths = append(subPaths, normRoot+","+childName)
		} else if childNodeType == "cmd" {
			hasCommandsDirectly = true
		}
	}

	// If root menu itself contains commands and no subdirs (e.g. "ip,address"), treat rootMenu as a submenu
	if len(subPaths) == 0 && hasCommandsDirectly {
		subPaths = append(subPaths, normRoot)
	}

	var submenus []*SubmenuSchema

	for _, subPath := range subPaths {
		parts := strings.Split(subPath, ",")
		lastPart := parts[len(parts)-1]

		schema := &SubmenuSchema{
			Name:     lastPart,
			FullPath: subPath,
			Commands: make(map[string]Command),
		}

		// List commands under this submenu
		cmdChildren, err := l.postInspect(ctx, map[string]string{
			"request": "child",
			"path":    subPath,
		})
		if err != nil {
			continue
		}

		availableCmds := make(map[string]bool)
		for _, cc := range cmdChildren {
			ccType, _ := cc["type"].(string)
			if ccType == "self" {
				continue
			}
			cmdName, _ := cc["name"].(string)
			cmdNodeType, _ := cc["node-type"].(string)
			if cmdNodeType == "cmd" {
				availableCmds[cmdName] = true
			}
		}

		// 1. Inspect 'add' command if available
		if availableCmds["add"] {
			addCmd := l.inspectCommand(ctx, subPath, "add")
			schema.Commands["add"] = addCmd
		}

		// 2. Inspect 'set' command if available
		if availableCmds["set"] {
			setCmd := l.inspectCommand(ctx, subPath, "set")
			schema.Commands["set"] = setCmd
		}

		// 3. Inspect 'print' (entity fields) if available
		if availableCmds["print"] {
			printCmd := l.inspectPrintEntity(ctx, subPath, schema)
			schema.Commands["print"] = printCmd
		}

		if len(schema.Commands) > 0 {
			submenus = append(submenus, schema)
		}
	}

	return submenus, nil
}

func (l *LiveRESTInspector) inspectCommand(ctx context.Context, subPath, cmdName string) Command {
	cmdPath := subPath + "," + cmdName
	cmd := Command{
		Name:     cmdName,
		FullPath: cmdPath,
	}

	syntaxRows, err := l.postInspect(ctx, map[string]string{
		"request": "syntax",
		"path":    cmdPath,
	})
	if err != nil {
		return cmd
	}

	seen := make(map[string]bool)
	for _, row := range syntaxRows {
		symbol, _ := row["symbol"].(string)
		if symbol == "" {
			symbol, _ = row["name"].(string)
		}

		if symbol == "<numbers>" {
			cmd.HasNumbers = true
			continue
		}

		if symbol == "" || symbol == "<value>" || strings.HasPrefix(symbol, "<") {
			continue
		}

		if seen[symbol] {
			continue
		}
		seen[symbol] = true

		helpText, _ := row["text"].(string)
		if helpText == "" {
			helpText, _ = row["help"].(string)
		}

		argType, _ := row["syntax-type"].(string)
		if argType == "" {
			argType = inferType(symbol, helpText)
		}

		isRequired := false
		if cmdName == "add" && symbol == "name" {
			isRequired = true
		}

		cmd.Arguments = append(cmd.Arguments, Argument{
			Name:        symbol,
			Type:        argType,
			Description: helpText,
			Required:    isRequired,
		})
	}

	return cmd
}

func (l *LiveRESTInspector) inspectPrintEntity(ctx context.Context, subPath string, schema *SubmenuSchema) Command {
	printCmd := Command{
		Name:     "print",
		FullPath: subPath + ",print",
	}

	cliPath := "/" + strings.ReplaceAll(subPath, ",", " ")
	complRows, err := l.postInspect(ctx, map[string]string{
		"request": "completion",
		"input":   cliPath + " print where ",
	})

	var propArgs []Argument
	var hasID bool

	if err == nil {
		seen := make(map[string]bool)
		for _, row := range complRows {
			style, _ := row["style"].(string)
			if style != "variable-local" {
				continue
			}

			propName, _ := row["completion"].(string)
			if propName == "" || propName == ".dead" || propName == ".nextid" ||
				propName == "about" || propName == "<value>" || strings.HasPrefix(propName, "<") {
				continue
			}

			if propName == ".id" {
				hasID = true
				continue
			}

			if seen[propName] {
				continue
			}
			seen[propName] = true

			helpText, _ := row["text"].(string)
			propArgs = append(propArgs, Argument{
				Name:        propName,
				Type:        inferType(propName, helpText),
				Description: helpText,
			})
		}
	}

	// Always place .id first if present
	if hasID {
		printCmd.Arguments = append(printCmd.Arguments, Argument{
			Name:        ".id",
			Type:        "string",
			Description: "Internal object identifier",
		})
	}
	printCmd.Arguments = append(printCmd.Arguments, propArgs...)

	// Fallback for singleton configuration menus (e.g. /ppp/aaa where "print where" does not apply)
	if len(printCmd.Arguments) == 0 {
		if setCmd, ok := schema.Commands["set"]; ok {
			for _, arg := range setCmd.Arguments {
				printCmd.Arguments = append(printCmd.Arguments, Argument{
					Name:        arg.Name,
					Type:        arg.Type,
					Description: arg.Description,
				})
			}
		} else if addCmd, ok := schema.Commands["add"]; ok {
			for _, arg := range addCmd.Arguments {
				printCmd.Arguments = append(printCmd.Arguments, Argument{
					Name:        arg.Name,
					Type:        arg.Type,
					Description: arg.Description,
				})
			}
		}
	}

	return printCmd
}

func inferType(name, desc string) string {
	lowerName := strings.ToLower(name)
	lowerDesc := strings.ToLower(desc)

	// Explicit 3-value enum fields in RouterOS that include "default"
	if lowerName == "use-mpls" || lowerName == "use-compression" || lowerName == "use-encryption" ||
		lowerName == "use-upnp" || lowerName == "only-one" || lowerName == "change-tcp-mss" {
		return "string"
	}

	switch {
	case lowerName == "disabled" || lowerName == "radius" || lowerName == "accounting" ||
		strings.HasPrefix(lowerName, "use-") || strings.HasPrefix(lowerName, "enable-") ||
		strings.HasPrefix(lowerName, "is-") || strings.Contains(lowerDesc, "enable/disable"):
		return "bool"
	case strings.Contains(lowerName, "bytes") || strings.Contains(lowerName, "quota") ||
		strings.Contains(lowerName, "packets") || strings.Contains(lowerName, "count"):
		return "int64"
	case lowerName == "mtu" || lowerName == "port" || strings.Contains(lowerName, "priority") ||
		strings.Contains(lowerName, "cost") || strings.Contains(lowerName, "vid"):
		return "int"
	case strings.HasSuffix(lowerName, "-list") || lowerName == "routes" || lowerName == "ipv6-routes":
		return "string"
	default:
		return "string"
	}
}

// JSONFileInspector loads a pre-dumped inspect JSON structure.
type JSONFileInspector struct {
	FilePath string
}

func (j *JSONFileInspector) InspectMenu(_ context.Context, _ string) ([]*SubmenuSchema, error) {
	data, err := os.ReadFile(j.FilePath)
	if err != nil {
		return nil, err
	}
	var schemas []*SubmenuSchema
	if err := json.Unmarshal(data, &schemas); err != nil {
		return nil, fmt.Errorf("unmarshal inspect file %s: %w", j.FilePath, err)
	}
	return schemas, nil
}

// EmbeddedInspector provides accurate pre-recorded RouterOS v7 inspect data for common menus (e.g. ppp).
type EmbeddedInspector struct{}

func (e *EmbeddedInspector) InspectMenu(_ context.Context, rootMenu string) ([]*SubmenuSchema, error) {
	norm := normalizePath(rootMenu)
	if norm != "ppp" {
		return nil, fmt.Errorf("embedded data only available for 'ppp' (for '%s', specify --host or --input)", rootMenu)
	}
	return getEmbeddedPPPSchema(), nil
}

func normalizePath(p string) string {
	p = strings.TrimSpace(p)
	p = strings.TrimPrefix(p, "/")
	p = strings.TrimSuffix(p, "/")
	p = strings.ReplaceAll(p, "/", ",")
	p = strings.ReplaceAll(p, " ", ",")
	parts := strings.Split(p, ",")
	var clean []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			clean = append(clean, part)
		}
	}
	return strings.Join(clean, ",")
}

// getEmbeddedPPPSchema returns the exact RouterOS v7 inspect definition for /ppp.
func getEmbeddedPPPSchema() []*SubmenuSchema {
	return []*SubmenuSchema{
		{
			Name:        "secret",
			FullPath:    "ppp,secret",
			Description: "PPP user secrets database for PPPoE, L2TP, PPTP, SSTP, and OVPN clients",
			Commands: map[string]Command{
				"print": {
					Name:     "print",
					FullPath: "ppp,secret,print",
					Arguments: []Argument{
						{Name: ".id", Type: "*HEX", Description: "Internal object identifier"},
						{Name: "name", Type: "string", Description: "Username for PPP authentication"},
						{Name: "password", Type: "string", Description: "Password for PPP authentication"},
						{Name: "service", Type: "string", Description: "PPP service type: any, async, l2tp, ovpn, pppoe, pptp, sstp"},
						{Name: "profile", Type: "string", Description: "Default profile for this user"},
						{Name: "local-address", Type: "ip", Description: "IP address assigned to server interface"},
						{Name: "remote-address", Type: "ip", Description: "IP address assigned to connecting client"},
						{Name: "caller-id", Type: "string", Description: "Caller ID / MAC address binding"},
						{Name: "routes", Type: "string", Description: "Static routes installed when secret connects"},
						{Name: "ipv6-routes", Type: "string", Description: "IPv6 routes installed when secret connects"},
						{Name: "limit-bytes-in", Type: "num", Description: "Transfer quota inbound in bytes"},
						{Name: "limit-bytes-out", Type: "num", Description: "Transfer quota outbound in bytes"},
						{Name: "last-logged-out", Type: "time", Description: "Timestamp of last logout"},
						{Name: "comment", Type: "string", Description: "Comment string"},
						{Name: "disabled", Type: "bool", Description: "Whether secret is disabled"},
					},
				},
				"add": {
					Name:     "add",
					FullPath: "ppp,secret,add",
					Arguments: []Argument{
						{Name: "name", Type: "string", Required: true, Description: "Username for PPP authentication"},
						{Name: "password", Type: "string", Description: "Password for PPP authentication"},
						{Name: "profile", Type: "string", Default: "default", Description: "Profile to assign to this secret"},
						{Name: "service", Type: "string", Default: "any", Description: "Service type: any, async, l2tp, ovpn, pppoe, pptp, sstp"},
						{Name: "local-address", Type: "ip", Description: "Server address for tunnel endpoint"},
						{Name: "remote-address", Type: "ip", Description: "Client address or pool name"},
						{Name: "caller-id", Type: "string", Description: "Bind user to caller ID or client MAC"},
						{Name: "routes", Type: "string", Description: "Routes added upon connection"},
						{Name: "limit-bytes-in", Type: "num", Description: "Maximum bytes downloaded before cutoff"},
						{Name: "limit-bytes-out", Type: "num", Description: "Maximum bytes uploaded before cutoff"},
						{Name: "comment", Type: "string", Description: "Descriptive comment"},
						{Name: "disabled", Type: "bool", Default: "no", Description: "Whether secret is disabled initially"},
					},
				},
				"set": {
					Name:     "set",
					FullPath: "ppp,secret,set",
					Arguments: []Argument{
						{Name: ".id", Type: "*HEX", Required: true, Description: "Internal ID or selector of target secret"},
						{Name: "name", Type: "string", Description: "Username for PPP authentication"},
						{Name: "password", Type: "string", Description: "Password for PPP authentication"},
						{Name: "profile", Type: "string", Description: "Profile to assign to this secret"},
						{Name: "service", Type: "string", Description: "Service type: any, async, l2tp, ovpn, pppoe, pptp, sstp"},
						{Name: "local-address", Type: "ip", Description: "Server address for tunnel endpoint"},
						{Name: "remote-address", Type: "ip", Description: "Client address or pool name"},
						{Name: "caller-id", Type: "string", Description: "Bind user to caller ID or client MAC"},
						{Name: "routes", Type: "string", Description: "Routes added upon connection"},
						{Name: "limit-bytes-in", Type: "num", Description: "Maximum bytes downloaded before cutoff"},
						{Name: "limit-bytes-out", Type: "num", Description: "Maximum bytes uploaded before cutoff"},
						{Name: "comment", Type: "string", Description: "Descriptive comment"},
						{Name: "disabled", Type: "bool", Description: "Whether secret is disabled"},
					},
				},
			},
		},
		{
			Name:        "profile",
			FullPath:    "ppp,profile",
			Description: "PPP user profiles defining bandwidth, queues, IP pools, and script hooks",
			Commands: map[string]Command{
				"print": {
					Name:     "print",
					FullPath: "ppp,profile,print",
					Arguments: []Argument{
						{Name: ".id", Type: "*HEX", Description: "Internal object identifier"},
						{Name: "name", Type: "string", Description: "Profile name"},
						{Name: "local-address", Type: "string", Description: "Default local address or pool"},
						{Name: "remote-address", Type: "string", Description: "Default remote address or pool"},
						{Name: "bridge", Type: "string", Description: "Bridge interface to bind customer sessions"},
						{Name: "rate-limit", Type: "string", Description: "Bandwidth limit in format rx/tx [burst-rx/burst-tx ...]"},
						{Name: "dns-server", Type: "string", Description: "DNS servers provided to clients"},
						{Name: "wins-server", Type: "string", Description: "WINS servers provided to clients"},
						{Name: "session-timeout", Type: "time", Description: "Maximum session lifetime"},
						{Name: "idle-timeout", Type: "time", Description: "Inactivity timeout before disconnect"},
						{Name: "use-mpls", Type: "string", Description: "Use MPLS: yes, no, or default"},
						{Name: "use-compression", Type: "string", Description: "Use packet compression"},
						{Name: "use-encryption", Type: "string", Description: "Use encryption: yes, no, or required"},
						{Name: "use-upnp", Type: "string", Description: "Allow UPnP for client sessions"},
						{Name: "only-one", Type: "string", Description: "Allow only one concurrent session per secret"},
						{Name: "change-tcp-mss", Type: "string", Description: "Automatically adjust TCP MSS on PPP interfaces"},
						{Name: "address-list", Type: "string", Description: "Firewall address list to add client IP to"},
						{Name: "parent-queue", Type: "string", Description: "Parent simple queue for client rate limits"},
						{Name: "on-up", Type: "string", Description: "Script executed when connection is established"},
						{Name: "on-down", Type: "string", Description: "Script executed when connection terminates"},
						{Name: "comment", Type: "string", Description: "Comment string"},
					},
				},
				"add": {
					Name:     "add",
					FullPath: "ppp,profile,add",
					Arguments: []Argument{
						{Name: "name", Type: "string", Required: true, Description: "Profile name"},
						{Name: "local-address", Type: "string", Description: "Default local address or pool"},
						{Name: "remote-address", Type: "string", Description: "Default remote address or pool"},
						{Name: "bridge", Type: "string", Description: "Bridge interface to bind customer sessions"},
						{Name: "rate-limit", Type: "string", Description: "Bandwidth limit (e.g. 10M/20M)"},
						{Name: "dns-server", Type: "string", Description: "DNS servers provided to clients"},
						{Name: "wins-server", Type: "string", Description: "WINS servers provided to clients"},
						{Name: "session-timeout", Type: "time", Description: "Maximum session lifetime"},
						{Name: "idle-timeout", Type: "time", Description: "Inactivity timeout before disconnect"},
						{Name: "use-mpls", Type: "string", Default: "default", Description: "Use MPLS: yes, no, or default"},
						{Name: "use-compression", Type: "string", Default: "default", Description: "Use packet compression"},
						{Name: "use-encryption", Type: "string", Default: "default", Description: "Use encryption: yes, no, or required"},
						{Name: "use-upnp", Type: "string", Default: "default", Description: "Allow UPnP for client sessions"},
						{Name: "only-one", Type: "string", Default: "default", Description: "Allow only one concurrent session"},
						{Name: "change-tcp-mss", Type: "string", Default: "default", Description: "Adjust TCP MSS"},
						{Name: "address-list", Type: "string", Description: "Firewall address list to add client IP to"},
						{Name: "parent-queue", Type: "string", Description: "Parent simple queue"},
						{Name: "on-up", Type: "string", Description: "Script hook on session start"},
						{Name: "on-down", Type: "string", Description: "Script hook on session teardown"},
						{Name: "comment", Type: "string", Description: "Descriptive comment"},
					},
				},
				"set": {
					Name:     "set",
					FullPath: "ppp,profile,set",
					Arguments: []Argument{
						{Name: ".id", Type: "*HEX", Required: true, Description: "Internal ID or selector of target profile"},
						{Name: "name", Type: "string", Description: "Profile name"},
						{Name: "local-address", Type: "string", Description: "Default local address or pool"},
						{Name: "remote-address", Type: "string", Description: "Default remote address or pool"},
						{Name: "bridge", Type: "string", Description: "Bridge interface to bind customer sessions"},
						{Name: "rate-limit", Type: "string", Description: "Bandwidth limit (e.g. 10M/20M)"},
						{Name: "dns-server", Type: "string", Description: "DNS servers provided to clients"},
						{Name: "wins-server", Type: "string", Description: "WINS servers provided to clients"},
						{Name: "session-timeout", Type: "time", Description: "Maximum session lifetime"},
						{Name: "idle-timeout", Type: "time", Description: "Inactivity timeout before disconnect"},
						{Name: "use-mpls", Type: "string", Description: "Use MPLS"},
						{Name: "use-compression", Type: "string", Description: "Use packet compression"},
						{Name: "use-encryption", Type: "string", Description: "Use encryption"},
						{Name: "use-upnp", Type: "string", Description: "Allow UPnP"},
						{Name: "only-one", Type: "string", Description: "Allow only one concurrent session"},
						{Name: "change-tcp-mss", Type: "string", Description: "Adjust TCP MSS"},
						{Name: "address-list", Type: "string", Description: "Firewall address list"},
						{Name: "parent-queue", Type: "string", Description: "Parent simple queue"},
						{Name: "on-up", Type: "string", Description: "Script hook on session start"},
						{Name: "on-down", Type: "string", Description: "Script hook on session teardown"},
						{Name: "comment", Type: "string", Description: "Descriptive comment"},
					},
				},
			},
		},
		{
			Name:        "active",
			FullPath:    "ppp,active",
			Description: "Active PPP client sessions currently connected to this router",
			Commands: map[string]Command{
				"print": {
					Name:     "print",
					FullPath: "ppp,active,print",
					Arguments: []Argument{
						{Name: ".id", Type: "*HEX", Description: "Internal session identifier"},
						{Name: "name", Type: "string", Description: "Username of connected PPP secret"},
						{Name: "service", Type: "string", Description: "Active connection type (pppoe, l2tp, sstp, etc.)"},
						{Name: "caller-id", Type: "string", Description: "Caller MAC address or remote IP"},
						{Name: "address", Type: "ip", Description: "IP address assigned to active client"},
						{Name: "uptime", Type: "time", Description: "Session connection duration"},
						{Name: "encoding", Type: "string", Description: "Cipher / protocol encoding in use"},
						{Name: "session-id", Type: "string", Description: "RADIUS session ID or internal session tag"},
						{Name: "limit-bytes-in", Type: "num", Description: "Remaining download quota"},
						{Name: "limit-bytes-out", Type: "num", Description: "Remaining upload quota"},
						{Name: "radius", Type: "bool", Description: "Whether session was authenticated via RADIUS"},
					},
				},
			},
		},
		{
			Name:        "aaa",
			FullPath:    "ppp,aaa",
			Description: "PPP Authentication, Authorization, and Accounting global configuration",
			Commands: map[string]Command{
				"print": {
					Name:     "print",
					FullPath: "ppp,aaa,print",
					Arguments: []Argument{
						{Name: "use-radius", Type: "bool", Description: "Authenticate PPP clients via RADIUS"},
						{Name: "accounting", Type: "bool", Description: "Send RADIUS accounting packets for active sessions"},
						{Name: "interim-update", Type: "time", Description: "Interval between periodic RADIUS interim updates"},
						{Name: "use-circuit-id", Type: "bool", Description: "Include DHCP Option 82 circuit-id in RADIUS request"},
					},
				},
				"set": {
					Name:     "set",
					FullPath: "ppp,aaa,set",
					Arguments: []Argument{
						{Name: "use-radius", Type: "bool", Description: "Authenticate PPP clients via RADIUS"},
						{Name: "accounting", Type: "bool", Description: "Send RADIUS accounting packets"},
						{Name: "interim-update", Type: "time", Description: "Interval between periodic RADIUS interim updates"},
						{Name: "use-circuit-id", Type: "bool", Description: "Include circuit-id in RADIUS request"},
					},
				},
			},
		},
	}
}
