package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	var (
		menuFlag     = flag.String("menu", "ppp", "RouterOS root menu hierarchy to inspect (e.g. 'ppp', 'ip/address', 'interface')")
		pkgFlag      = flag.String("pkg", "", "Go package name for the generated code (default: derived from menu)")
		outFlag      = flag.String("out", "", "Output Go file path (default: <menu>.go, use '-' for stdout)")
		hostFlag     = flag.String("host", "", "RouterOS device IP/hostname for live /console/inspect querying")
		portFlag     = flag.Int("port", 80, "RouterOS REST API port (usually 80 or 443)")
		userFlag     = flag.String("user", "admin", "RouterOS username")
		passFlag     = flag.String("pass", "", "RouterOS password")
		useTLSFlag   = flag.Bool("use-tls", false, "Use HTTPS for REST API")
		insecureFlag = flag.Bool("insecure", false, "Skip TLS certificate verification")
		inputFlag    = flag.String("input", "", "Path to pre-dumped inspect JSON file (offline mode)")
		dumpFlag     = flag.String("dump", "", "Save introspected schemas to JSON file for caching/offline use")
	)
	flag.Parse()

	rootMenu := strings.TrimSpace(*menuFlag)
	if rootMenu == "" {
		fmt.Fprintf(os.Stderr, "Error: --menu is required\n")
		os.Exit(1)
	}

	pkgName := *pkgFlag
	if pkgName == "" {
		// Use the last segment of the menu, e.g. "ip/address" -> "address", "ppp" -> "ppp"
		parts := strings.FieldsFunc(rootMenu, func(r rune) bool {
			return r == '/' || r == ','
		})
		if len(parts) > 0 {
			pkgName = strings.ToLower(parts[len(parts)-1])
		} else {
			pkgName = "mikrotik"
		}
	}

	outFile := *outFlag
	if outFile == "" {
		baseName := strings.ReplaceAll(rootMenu, "/", "_")
		baseName = strings.ReplaceAll(baseName, ",", "_")
		outFile = baseName + ".go"
	}

	var inspector Inspector
	switch {
	case *hostFlag != "":
		fmt.Printf("Connecting to RouterOS live endpoint at %s:%d (/console/inspect)...\n", *hostFlag, *portFlag)
		inspector = NewLiveRESTInspector(*hostFlag, *portFlag, *userFlag, *passFlag, *useTLSFlag, *insecureFlag)
	case *inputFlag != "":
		fmt.Printf("Loading inspect schema from file: %s...\n", *inputFlag)
		inspector = &JSONFileInspector{FilePath: *inputFlag}
	default:
		fmt.Printf("No --host or --input specified. Using RouterOS v7 inspect catalog for '/%s'...\n", rootMenu)
		inspector = &EmbeddedInspector{}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	schemas, err := inspector.InspectMenu(ctx, rootMenu)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Inspection failed: %v\n", err)
		os.Exit(1)
	}

	if len(schemas) == 0 {
		fmt.Fprintf(os.Stderr, "No submenus or commands found under /%s\n", rootMenu)
		os.Exit(1)
	}

	// Optionally dump raw schemas to JSON
	if *dumpFlag != "" {
		dumpBytes, err := json.MarshalIndent(schemas, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to marshal inspect dump: %v\n", err)
		} else if err := os.WriteFile(*dumpFlag, dumpBytes, 0o600); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write dump to %s: %v\n", *dumpFlag, err)
		} else {
			fmt.Printf("Saved inspect schema dump to %s\n", *dumpFlag)
		}
	}

	// Generate Go source code
	opts := GeneratorOptions{
		PackageName: pkgName,
		RootMenu:    rootMenu,
	}

	sourceBytes, err := GenerateSource(schemas, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Generation error: %v\n", err)
		os.Exit(1)
	}

	if outFile == "-" {
		fmt.Print(string(sourceBytes))
		return
	}

	// Ensure target directory exists
	dir := filepath.Dir(outFile)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create directory %s: %v\n", dir, err)
			os.Exit(1)
		}
	}

	if err := os.WriteFile(outFile, sourceBytes, 0o600); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write output to %s: %v\n", outFile, err)
		os.Exit(1)
	}

	fmt.Printf("Successfully generated %s (%d bytes)\n", outFile, len(sourceBytes))
	fmt.Printf("Submenus processed:\n")
	for _, s := range schemas {
		fmt.Printf("  • /%s (commands: ", strings.ReplaceAll(s.FullPath, ",", "/"))
		var cmds []string
		for cmdName := range s.Commands {
			cmds = append(cmds, cmdName)
		}
		fmt.Printf("%s)\n", strings.Join(cmds, ", "))
	}
}
