package wifi

import (
	"testing"
)

func TestCAPsMANCommandsAndParsers(t *testing.T) {
	mgrCmd := NewPrintCAPsMANManagerCommand()
	if mgrCmd.Raw != "/interface/wifi/capsman/print" || mgrCmd.Args[".proplist"] == "" {
		t.Fatalf("invalid capsman print: %v", mgrCmd)
	}

	capsCmd := NewPrintRemoteCAPsCommand()
	if capsCmd.Raw != "/interface/wifi/capsman/remote-cap/print" {
		t.Fatalf("invalid remote caps print: %v", capsCmd)
	}

	mgr := ParseCAPsMANManager([]map[string]string{
		{"enabled": "yes", "interfaces": "bridge", "upgrade-policy": "suggest"},
	})
	if mgr == nil || !mgr.Enabled || mgr.Interfaces != "bridge" {
		t.Errorf("unexpected capsman manager parse: %+v", mgr)
	}

	caps := ParseRemoteCAPs([]map[string]string{
		{".id": "*1", "identity": "AP-Lobby", "board-name": "cAP ax", "state": "running"},
	})
	if len(caps) != 1 || caps[0].Identity != "AP-Lobby" {
		t.Errorf("unexpected caps parse: %+v", caps)
	}

	datapaths := ParseCAPsDatapaths([]map[string]string{
		{".id": "*1", "name": "dp-guest", "bridge": "bridge", "vlan-id": "20", "local-forwarding": "yes"},
	})
	if len(datapaths) != 1 || datapaths[0].VLANID != 20 || !datapaths[0].LocalForwarding {
		t.Errorf("unexpected datapaths parse: %+v", datapaths)
	}
}

func TestWiFiScanCommandsAndParsers(t *testing.T) {
	scanCmd := NewWiFiFrequencyScanCommand("wifi1", "5s")
	if scanCmd.Raw != "/interface/wifi/frequency-scan" || scanCmd.Args["duration"] != "5s" {
		t.Fatalf("invalid wifi scan cmd: %v", scanCmd)
	}

	results := ParseScanResults([]map[string]string{
		{"ssid": "Office-Guest", "bssid": "00:11:22:33:44:55", "frequency": "5180", "signal": "-65"},
	})
	if len(results) != 1 || results[0].SSID != "Office-Guest" || results[0].Frequency != 5180 || results[0].Signal != -65 {
		t.Errorf("unexpected scan results parse: %+v", results)
	}
}

func TestLegacyWirelessSecurityProfile(t *testing.T) {
	printProfiles := NewPrintWirelessSecurityProfilesCommand()
	if printProfiles.Raw != "/interface/wireless/security-profiles/print" || printProfiles.Args[".proplist"] == "" {
		t.Fatalf("invalid print profiles: %v", printProfiles)
	}

	addProf := NewAddWirelessSecurityProfileCommand(WirelessSecurityProfile{
		Name:                "profile-wpa2",
		Mode:                "dynamic-keys",
		AuthenticationTypes: "wpa2-psk",
		UnicastCiphers:      "aes-ccm",
		GroupCiphers:        "aes-ccm",
		Comment:             "Office WiFi",
	}, "", "supersecretpass")
	if addProf.Raw != "/interface/wireless/security-profiles/add" || addProf.Args["wpa2-pre-shared-key"] != "supersecretpass" {
		t.Errorf("invalid add profile: %v", addProf)
	}

	setProf := NewSetWirelessSecurityProfileCommand("*1", WirelessSecurityProfile{
		Name: "profile-wpa2",
	}, "", "newpass")
	if setProf.Raw != "/interface/wireless/security-profiles/set" || setProf.Args[".id"] != "*1" {
		t.Errorf("invalid set profile: %v", setProf)
	}

	remProf := NewRemoveWirelessSecurityProfileCommand("*1")
	if remProf.Raw != "/interface/wireless/security-profiles/remove" || remProf.Args[".id"] != "*1" {
		t.Errorf("invalid rem profile: %v", remProf)
	}

	profiles := ParseWirelessSecurityProfiles([]map[string]string{
		{".id": "*1", "name": "profile-wpa2", "mode": "dynamic-keys", "authentication-types": "wpa2-psk"},
	})
	if len(profiles) != 1 || profiles[0].AuthenticationTypes != "wpa2-psk" {
		t.Errorf("unexpected profiles parse: %+v", profiles)
	}
}
