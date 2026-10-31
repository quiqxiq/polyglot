package system

import (
	"strings"
	"testing"
	"time"
)

func TestBackupCommands(t *testing.T) {
	saveCmd := NewBackupSaveCommand(BackupSaveConfig{
		Name:        "backup-01",
		Password:    "secret123",
		Encryption:  "aes-sha256",
		DontEncrypt: false,
	})
	if saveCmd.Raw != "/system/backup/save" {
		t.Fatalf("expected /system/backup/save, got %s", saveCmd.Raw)
	}
	if saveCmd.Args["name"] != "backup-01" || saveCmd.Args["encryption"] != "aes-sha256" {
		t.Errorf("unexpected save args: %v", saveCmd.Args)
	}

	loadCmd := NewBackupLoadCommand("backup-01", "secret123")
	if loadCmd.Raw != "/system/backup/load" {
		t.Fatalf("expected /system/backup/load, got %s", loadCmd.Raw)
	}

	expCmd := NewExportCommand(ExportConfig{File: "cfg.rsc", Terse: true, ShowSensitive: true})
	if expCmd.Raw != "/export" {
		t.Fatalf("expected /export, got %s", expCmd.Raw)
	}
	if _, ok := expCmd.Args["terse"]; !ok {
		t.Error("expected terse flag in export")
	}

	// Test Settled helper
	sizes := []int{100, 200, 200}
	idx := 0
	sz, err := Settled(func(name string) (int, error) {
		if idx < len(sizes) {
			s := sizes[idx]
			idx++
			return s, nil
		}
		return 200, nil
	}, "test.rsc", 2*time.Second, func(d time.Duration) {})
	if err != nil || sz != 200 {
		t.Errorf("Settled failed: sz=%d, err=%v", sz, err)
	}

	// Test Sweep helper
	var removed []string
	n, err := Sweep(func() ([]string, error) {
		return []string{"tmp-1", "tmp-2", "keep-me"}, nil
	}, func(name string) error {
		removed = append(removed, name)
		return nil
	}, "tmp-")
	if err != nil || n != 2 || len(removed) != 2 {
		t.Errorf("Sweep failed: n=%d, removed=%v, err=%v", n, removed, err)
	}
}

func TestFileCommandsAndParsers(t *testing.T) {
	listCmd := NewPrintFilesCommand()
	if listCmd.Raw != "/file/print" || listCmd.Args[".proplist"] == "" {
		t.Errorf("invalid file print cmd: %v", listCmd)
	}

	findCmd := NewPrintFileCommand("backup.rsc")
	if findCmd.Args["?name"] != "backup.rsc" {
		t.Errorf("invalid find cmd: %v", findCmd)
	}

	readChunkCmd := NewReadFileChunkCommand("file.txt", 0, 32768)
	if readChunkCmd.Raw != "/file/read" || readChunkCmd.Args["chunk-size"] != "32768" {
		t.Errorf("invalid read chunk cmd: %v", readChunkCmd)
	}

	// Secret masking
	raw := `add name=user1 password="mysecretpassword" comment=test secret=supersecret psk=12345678`
	masked, count := MaskSecrets(raw)
	if count != 3 || strings.Contains(masked, "mysecretpassword") || strings.Contains(masked, "supersecret") {
		t.Errorf("mask failed: count=%d, masked=%s", count, masked)
	}

	// File validation
	if err := ValidateTextFile([]byte("hello world"), 100); err != nil {
		t.Errorf("expected valid text file: %v", err)
	}
	if err := ValidateTextFile([]byte("-----BEGIN RSA PRIVATE KEY-----"), 100); err == nil {
		t.Error("expected error for private key")
	}

	// Parse file entries
	rows := []map[string]string{
		{".id": "*1", "name": "backup.rsc", "type": "script", "size": " 1024 ", "creation-time": "oct/28/2026 10:00:00"},
	}
	files := ParseFileEntries(rows)
	if len(files) != 1 || files[0].Size != 1024 || files[0].Name != "backup.rsc" {
		t.Errorf("unexpected files parse: %+v", files)
	}
}

func TestPackageCommandsAndParsers(t *testing.T) {
	pkgCmd := NewPrintPackagesCommand()
	if pkgCmd.Raw != "/system/package/print" {
		t.Errorf("invalid pkg print cmd: %v", pkgCmd)
	}

	schCmd, err := NewPackageScheduleCommand("uninstall", "*3")
	if err != nil || schCmd.Raw != "/system/package/uninstall" || schCmd.Args["numbers"] != "*3" {
		t.Errorf("invalid sch cmd: %v, err=%v", schCmd, err)
	}

	pkgs := ParsePackages([]map[string]string{
		{".id": "*1", "name": "routeros", "version": "7.24.4", "disabled": "false", "size": "25000000"},
	})
	if len(pkgs) != 1 || pkgs[0].Version != "7.24.4" || pkgs[0].Size != 25000000 {
		t.Errorf("unexpected pkgs parse: %+v", pkgs)
	}

	update := ParsePackageUpdate([]map[string]string{
		{"installed-version": "7.24.4", "latest-version": "7.25", "status": "New version available", "channel": "stable"},
	})
	if update == nil || update.LatestVersion != "7.25" {
		t.Errorf("unexpected update parse: %+v", update)
	}
}

func TestUserCommandsAndParsers(t *testing.T) {
	usrCmd := NewPrintUsersCommand()
	if usrCmd.Raw != "/user/print" {
		t.Errorf("invalid user print cmd: %v", usrCmd)
	}

	// Test user group set with NegateUnset
	grpCmd := NewSetGroupCommand(RosGroup{
		ID:     "*2",
		Name:   "operators",
		Policy: []string{"read", "api", "test"},
	})
	if grpCmd.Raw != "/user/group/set" {
		t.Fatalf("expected /user/group/set, got %s", grpCmd.Raw)
	}
	pols := strings.Split(grpCmd.Args["policy"], ",")
	hasNegatedWrite := false
	hasRead := false
	for _, p := range pols {
		if p == "!write" {
			hasNegatedWrite = true
		}
		if p == "read" {
			hasRead = true
		}
	}
	if !hasNegatedWrite || !hasRead {
		t.Errorf("NegateUnset policy failed: %s", grpCmd.Args["policy"])
	}

	users := ParseUsers([]map[string]string{
		{".id": "*1", "name": "admin", "group": "full", "disabled": "false"},
	})
	if len(users) != 1 || users[0].Name != "admin" {
		t.Errorf("unexpected users parse: %+v", users)
	}
}

func TestNTPAndLoggingCommandsAndParsers(t *testing.T) {
	ntpCmd := NewPrintNTPClientCommand()
	if ntpCmd.Raw != "/system/ntp/client/print" {
		t.Errorf("invalid ntp cmd: %v", ntpCmd)
	}
	ntp := ParseNTPClient([]map[string]string{
		{"enabled": "yes", "mode": "unicast", "servers": "time.google.com,pool.ntp.org", "status": "synchronized"},
	})
	if ntp == nil || !ntp.Enabled || len(ntp.Servers) != 2 {
		t.Errorf("unexpected ntp parse: %+v", ntp)
	}

	logCmd := NewPrintLogRulesCommand()
	if logCmd.Raw != "/system/logging/print" {
		t.Errorf("invalid log rule cmd: %v", logCmd)
	}
	rules := ParseLogRules([]map[string]string{
		{".id": "*1", "topics": "info,account", "action": "memory", "disabled": "false"},
	})
	if len(rules) != 1 || rules[0].Action != "memory" {
		t.Errorf("unexpected log rules parse: %+v", rules)
	}
}

func TestLifecycleCommandsAndParsers(t *testing.T) {
	rebootCmd := NewRebootCommand()
	if rebootCmd.Raw != "/system/reboot" {
		t.Errorf("invalid reboot cmd: %v", rebootCmd)
	}

	resetCmd := NewResetConfigurationCommand(ResetConfigParams{
		NoDefaults: true,
		SkipBackup: true,
	})
	if resetCmd.Raw != "/system/reset-configuration" || resetCmd.Args["no-defaults"] != "yes" {
		t.Errorf("invalid reset cmd: %v", resetCmd)
	}

	lic := ParseLicenseInfo([]map[string]string{
		{"system-id": "1234", "software-id": "ABCD-EFGH", "level": "6"},
	})
	if lic == nil || lic.Level != "6" {
		t.Errorf("unexpected lic parse: %+v", lic)
	}

	devMode := ParseDeviceMode([]map[string]string{
		{"mode": "enterprise", "container": "yes", "sniffer": "yes"},
	})
	if devMode == nil || !devMode.ContainerAllowed {
		t.Errorf("unexpected devMode parse: %+v", devMode)
	}
}

func TestScriptAndSchedulerCommandsAndParsers(t *testing.T) {
	scriptPrint := NewPrintScriptCommand("")
	if scriptPrint.Raw != "/system/script/print" || scriptPrint.Args[".proplist"] == "" {
		t.Errorf("invalid script print cmd: %v", scriptPrint)
	}

	addScript := NewAddScriptCommand(SystemScriptParams{
		Name:       "test-script",
		Source:     ":log info \"hello\"",
		DontReqPwr: true,
	})
	if addScript.Raw != "/system/script/add" || addScript.Args["dont-require-permissions"] != "yes" {
		t.Errorf("invalid add script: %v", addScript)
	}

	runScript := NewRunScriptCommand("test-script")
	if runScript.Raw != "/system/script/run" || runScript.Args["number"] != "test-script" {
		t.Errorf("invalid run script: %v", runScript)
	}

	schedPrint := NewPrintSchedulerCommand("")
	if schedPrint.Raw != "/system/scheduler/print" {
		t.Errorf("invalid sched print cmd: %v", schedPrint)
	}

	addSched := NewAddSchedulerCommand(SystemSchedulerParams{
		Name:     "nightly-backup",
		Interval: "1d",
		OnEvent:  "/system backup save",
		Disabled: false,
	})
	if addSched.Raw != "/system/scheduler/add" || addSched.Args["interval"] != "1d" {
		t.Errorf("invalid add sched: %v", addSched)
	}

	ntpSrvPrint := NewPrintNTPServerSettingsCommand()
	if ntpSrvPrint.Raw != "/system/ntp/server/print" {
		t.Errorf("invalid ntp server print: %v", ntpSrvPrint)
	}

	ntpSrv := ParseNTPServerSettings([]map[string]string{
		{"enabled": "yes", "manycast": "no", "multicast": "yes", "broadcast": "no"},
	})
	if ntpSrv == nil || !ntpSrv.Enabled || !ntpSrv.Multicast {
		t.Errorf("unexpected ntp server parse: %+v", ntpSrv)
	}
}
