package mikrotik

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/port"
)

type cacheEntry struct {
	data      any
	expiresAt time.Time
}

// InspectCache provides in-memory thread-safe caching for static inspect queries (child, syntax).
type InspectCache struct {
	mu      sync.RWMutex
	entries map[string]cacheEntry
	ttl     time.Duration
}

// NewInspectCache creates an InspectCache with the given TTL (defaults to 30m if non-positive).
func NewInspectCache(ttl time.Duration) *InspectCache {
	if ttl <= 0 {
		ttl = 30 * time.Minute
	}
	return &InspectCache{
		entries: make(map[string]cacheEntry),
		ttl:     ttl,
	}
}

// Get retrieves cached data for the key if present and unexpired.
func (c *InspectCache) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	e, ok := c.entries[key]
	if !ok || time.Now().After(e.expiresAt) {
		return nil, false
	}
	return e.data, true
}

// Set stores data in the cache with the configured TTL.
func (c *InspectCache) Set(key string, data any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = cacheEntry{
		data:      data,
		expiresAt: time.Now().Add(c.ttl),
	}
}

// Clear flushes all entries from the cache.
func (c *InspectCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]cacheEntry)
}

// InspectClient coordinates /console/inspect queries over RouterOS binary API.
type InspectClient struct {
	cache *InspectCache
}

// NewInspectClient constructs a new InspectClient with default caching.
func NewInspectClient() *InspectClient {
	return &InspectClient{
		cache: NewInspectCache(30 * time.Minute),
	}
}

// ListMenu lists submenus, commands, and arguments under a RouterOS hierarchy path.
// Static results are cached by deviceID and path.
func (c *InspectClient) ListMenu(ctx context.Context, driver port.DeviceDriver, deviceID, rawPath string) ([]MenuChildItem, bool, error) {
	path := normalizePath(rawPath)
	cacheKey := fmt.Sprintf("%s:child:%s", deviceID, path)

	if cached, ok := c.cache.Get(cacheKey); ok {
		if items, ok := cached.([]MenuChildItem); ok {
			return items, true, nil
		}
	}

	cmdArgs := map[string]string{
		"request": "child",
	}
	if path != "" {
		cmdArgs["path"] = path
	}

	res, err := driver.Execute(ctx, command.Command{
		Raw:  "/console/inspect",
		Args: cmdArgs,
	})
	if err != nil {
		return nil, false, fmt.Errorf("inspect child failed: %w", err)
	}

	items := make([]MenuChildItem, 0, len(res.Rows))
	for _, row := range res.Rows {
		name := row["name"]
		if name == "" {
			continue
		}
		item := MenuChildItem{
			Name:     name,
			Type:     row["type"],
			NodeType: row["node-type"],
			Help:     row["help"],
		}
		if item.NodeType == "" {
			item.NodeType = row["node_type"]
		}
		if item.Help == "" {
			item.Help = row["description"]
		}
		items = append(items, item)
	}

	c.cache.Set(cacheKey, items)
	return items, false, nil
}

// DescribeCommand returns syntax and argument information for a given command path.
// Static results are cached by deviceID and path.
func (c *InspectClient) DescribeCommand(ctx context.Context, driver port.DeviceDriver, deviceID, rawPath string) ([]CommandArgument, bool, error) {
	path := normalizePath(rawPath)
	if path == "" {
		return nil, false, fmt.Errorf("command path cannot be empty")
	}
	cacheKey := fmt.Sprintf("%s:describe:%s", deviceID, path)

	if cached, ok := c.cache.Get(cacheKey); ok {
		if args, ok := cached.([]CommandArgument); ok {
			return args, true, nil
		}
	}

	// First query syntax on the command path
	syntaxRes, err := driver.Execute(ctx, command.Command{
		Raw: "/console/inspect",
		Args: map[string]string{
			"request": "syntax",
			"path":    path,
		},
	})
	if err != nil {
		return nil, false, fmt.Errorf("inspect syntax failed: %w", err)
	}

	var args []CommandArgument
	seen := make(map[string]bool)

	for _, row := range syntaxRes.Rows {
		name := row["name"]
		if name == "" {
			continue
		}
		seen[name] = true
		args = append(args, parseCommandArgument(row))
	}

	// Query child elements to catch any arguments not detailed by syntax
	childRes, childErr := driver.Execute(ctx, command.Command{
		Raw: "/console/inspect",
		Args: map[string]string{
			"request": "child",
			"path":    path,
		},
	})
	if childErr == nil {
		for _, row := range childRes.Rows {
			name := row["name"]
			if name == "" || seen[name] {
				continue
			}
			typ := row["type"]
			if typ == "cmd" || typ == "dir" {
				continue
			}
			args = append(args, CommandArgument{
				Name:        name,
				Type:        typ,
				Description: row["help"],
			})
			seen[name] = true
		}
	}

	c.cache.Set(cacheKey, args)
	return args, false, nil
}

// Suggest queries dynamic completion suggestions for a console input string.
// Dynamic completions reflect live router state (interfaces, profiles) and are never cached.
func (c *InspectClient) Suggest(ctx context.Context, driver port.DeviceDriver, rawInput string) ([]SuggestionItem, error) {
	input := normalizeConsoleInput(rawInput)

	res, err := driver.Execute(ctx, command.Command{
		Raw: "/console/inspect",
		Args: map[string]string{
			"request": "completion",
			"input":   input,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("inspect completion failed: %w", err)
	}

	var suggestions []SuggestionItem
	for _, row := range res.Rows {
		if row["show"] == "false" {
			continue
		}
		comp := row["completion"]
		if comp == "" {
			continue
		}
		suggestions = append(suggestions, SuggestionItem{
			Completion: comp,
			Style:      row["style"],
			Preference: row["preference"],
		})
	}
	return suggestions, nil
}

// Validate dry-runs syntax and argument validation on a console input line without executing.
// Dynamic validation is never cached.
func (c *InspectClient) Validate(ctx context.Context, driver port.DeviceDriver, rawInput string) (ValidationResult, error) {
	input := normalizeConsoleInput(rawInput)

	res, err := driver.Execute(ctx, command.Command{
		Raw: "/console/inspect",
		Args: map[string]string{
			"request": "error",
			"input":   input,
		},
	})
	if err != nil {
		lowerErr := strings.ToLower(err.Error())
		if strings.Contains(lowerErr, "no such command") || strings.Contains(lowerErr, "unknown command") {
			// RouterOS v6 graceful degradation
			return ValidationResult{Valid: true}, nil
		}
		return ValidationResult{}, fmt.Errorf("inspect error failed: %w", err)
	}

	for _, row := range res.Rows {
		errMsg := row["error"]
		if errMsg != "" {
			pos := parsePosition(row)
			return ValidationResult{
				Valid:         false,
				ErrorMessage:  errMsg,
				ErrorPosition: pos,
			}, nil
		}
	}

	return ValidationResult{Valid: true}, nil
}

func parseCommandArgument(row map[string]string) CommandArgument {
	arg := CommandArgument{
		Name:        row["name"],
		Type:        row["type"],
		Default:     row["default"],
		Description: row["help"],
	}
	if arg.Description == "" {
		arg.Description = row["description"]
	}
	if req := strings.ToLower(row["required"]); req == "yes" || req == "true" {
		arg.Required = true
	}
	if vals := row["values"]; vals != "" {
		split := strings.FieldsFunc(vals, func(r rune) bool {
			return r == ',' || r == '|' || r == ' '
		})
		for _, v := range split {
			v = strings.TrimSpace(v)
			if v != "" {
				arg.Values = append(arg.Values, v)
			}
		}
	}
	return arg
}

func parsePosition(row map[string]string) int {
	for _, key := range []string{"offset", "position", "column"} {
		if val := row[key]; val != "" {
			if n, err := strconv.Atoi(val); err == nil {
				return n
			}
		}
	}
	return 0
}

// normalizePath converts various path forms (e.g. "/ip/address/add", "ip address add")
// into canonical comma-separated RouterOS inspect path ("ip,address,add").
func normalizePath(p string) string {
	p = strings.TrimSpace(p)
	p = strings.TrimPrefix(p, "/")
	p = strings.TrimSuffix(p, "/")
	if p == "" {
		return ""
	}
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

// normalizeConsoleInput ensures a console input line starts with "/" if formatted like a path.
func normalizeConsoleInput(in string) string {
	in = strings.TrimSpace(in)
	if in == "" {
		return "/"
	}
	if strings.HasPrefix(in, "/") {
		return in
	}
	// Check common root namespaces
	roots := []string{"ip ", "interface ", "routing ", "system ", "tool ", "certificate ", "queue ", "user "}
	for _, r := range roots {
		if strings.HasPrefix(in, r) {
			return "/" + in
		}
	}
	return in
}

// buildConsoleLine formats a command raw path and argument map into a single console line
// suitable for /console/inspect request=error or request=completion.
func buildConsoleLine(raw string, args map[string]string) string {
	raw = strings.TrimSpace(raw)
	if len(args) == 0 {
		if !strings.HasPrefix(raw, "/") {
			raw = "/" + raw
		}
		if !strings.Contains(raw, " ") {
			parts := strings.Split(strings.Trim(raw, "/"), "/")
			return "/" + strings.Join(parts, " ")
		}
		return raw
	}

	cleanPath := strings.Trim(raw, "/")
	parts := strings.Split(cleanPath, "/")
	base := "/" + strings.Join(parts, " ")

	keys := make([]string, 0, len(args))
	for k := range args {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	sb.WriteString(base)
	for _, k := range keys {
		v := args[k]
		sb.WriteString(" ")
		if v == "" {
			sb.WriteString(k)
		} else {
			sb.WriteString(k)
			sb.WriteString("=")
			if strings.Contains(v, " ") && !strings.HasPrefix(v, "\"") {
				sb.WriteString("\"" + v + "\"")
			} else {
				sb.WriteString(v)
			}
		}
	}
	return sb.String()
}

// parseConsoleCommand decomposes a console command string into a canonical RouterOS API path
// (e.g. "/ip/address/add") and an attribute argument map.
func parseConsoleCommand(input string) (string, map[string]string) {
	input = strings.TrimSpace(input)
	if !strings.Contains(input, " ") {
		if !strings.HasPrefix(input, "/") {
			input = "/" + input
		}
		return input, make(map[string]string)
	}

	tokens := tokenizeConsole(input)
	var pathParts []string
	args := make(map[string]string)
	parsingArgs := false

	for _, token := range tokens {
		if !parsingArgs {
			if strings.Contains(token, "=") {
				parsingArgs = true
			} else {
				clean := strings.TrimPrefix(token, "/")
				if clean != "" {
					pathParts = append(pathParts, clean)
				}
				continue
			}
		}

		if strings.Contains(token, "=") {
			pair := strings.SplitN(token, "=", 2)
			args[pair[0]] = pair[1]
		} else {
			args[token] = ""
		}
	}

	raw := "/" + strings.Join(pathParts, "/")
	return raw, args
}

// tokenizeConsole tokenizes a command line respecting quoted strings.
func tokenizeConsole(s string) []string {
	var tokens []string
	var cur strings.Builder
	inDouble := false
	inSingle := false

	for i := 0; i < len(s); i++ {
		ch := s[i]
		switch ch {
		case '"':
			if !inSingle {
				inDouble = !inDouble
				continue
			}
			cur.WriteByte(ch)
		case '\'':
			if !inDouble {
				inSingle = !inSingle
				continue
			}
			cur.WriteByte(ch)
		case ' ', '\t':
			if inDouble || inSingle {
				cur.WriteByte(ch)
			} else if cur.Len() > 0 {
				tokens = append(tokens, cur.String())
				cur.Reset()
			}
		default:
			cur.WriteByte(ch)
		}
	}
	if cur.Len() > 0 {
		tokens = append(tokens, cur.String())
	}
	return tokens
}
