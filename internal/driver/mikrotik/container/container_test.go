package container

import (
	"testing"
)

func TestContainerCommandsAndParsers(t *testing.T) {
	printCmd := NewPrintContainersCommand()
	if printCmd.Raw != "/container/print" || printCmd.Args[".proplist"] == "" {
		t.Fatalf("invalid container print: %v", printCmd)
	}

	startCmd := NewStartContainerCommand("*1")
	if startCmd.Raw != "/container/start" || startCmd.Args[".id"] != "*1" {
		t.Errorf("unexpected start cmd: %v", startCmd)
	}

	stopCmd := NewStopContainerCommand("*1")
	if stopCmd.Raw != "/container/stop" || stopCmd.Args[".id"] != "*1" {
		t.Errorf("unexpected stop cmd: %v", stopCmd)
	}

	appCmd := NewPrintAppsCommand()
	if appCmd.Raw != "/app/print" {
		t.Fatalf("invalid app print: %v", appCmd)
	}

	enableCmd := NewEnableAppCommand("*2")
	if enableCmd.Raw != "/app/enable" || enableCmd.Args["numbers"] != "*2" {
		t.Errorf("unexpected enable cmd: %v", enableCmd)
	}

	containers := ParseContainers([]map[string]string{
		{".id": "*1", "name": "pihole", "status": "running", "running": "true", "image": "pihole/pihole:latest"},
	})
	if len(containers) != 1 || !containers[0].Running || containers[0].Name != "pihole" {
		t.Errorf("unexpected containers parse: %+v", containers)
	}

	apps := ParseApps([]map[string]string{
		{".id": "*1", "name": "adguard-home", "running": "true", "app-size": "30000000"},
	})
	if len(apps) != 1 || apps[0].AppSize != 30000000 || !apps[0].Running {
		t.Errorf("unexpected apps parse: %+v", apps)
	}
}

func TestContainerEnvsMountsVeth(t *testing.T) {
	// Envs
	printEnvs := NewPrintContainerEnvsCommand("pihole_envs")
	if printEnvs.Raw != "/container/envs/print" || printEnvs.Args["?list"] != "pihole_envs" || printEnvs.Args[".proplist"] == "" {
		t.Errorf("invalid print envs: %v", printEnvs)
	}

	addEnv := NewAddContainerEnvCommand(ContainerEnv{
		List:  "pihole_envs",
		Key:   "TZ",
		Value: "Asia/Jakarta",
	})
	if addEnv.Raw != "/container/envs/add" || addEnv.Args["value"] != "Asia/Jakarta" {
		t.Errorf("invalid add env: %v", addEnv)
	}

	envs := ParseContainerEnvs([]map[string]string{
		{".id": "*1", "list": "pihole_envs", "key": "TZ"},
	})
	if len(envs) != 1 || envs[0].Key != "TZ" {
		t.Errorf("unexpected envs parse: %+v", envs)
	}

	// Mounts
	printMounts := NewPrintContainerMountsCommand()
	if printMounts.Raw != "/container/mounts/print" {
		t.Errorf("invalid print mounts: %v", printMounts)
	}

	addMount := NewAddContainerMountCommand(ContainerMount{
		Name: "etc-pihole",
		Src:  "/usb1/pihole/etc",
		Dst:  "/etc/pihole",
	})
	if addMount.Raw != "/container/mounts/add" || addMount.Args["dst"] != "/etc/pihole" {
		t.Errorf("invalid add mount: %v", addMount)
	}

	mounts := ParseContainerMounts([]map[string]string{
		{".id": "*1", "name": "etc-pihole", "src": "/usb1/pihole/etc", "dst": "/etc/pihole"},
	})
	if len(mounts) != 1 || mounts[0].Dst != "/etc/pihole" {
		t.Errorf("unexpected mounts parse: %+v", mounts)
	}

	// Veth
	printVeth := NewPrintVethInterfacesCommand()
	if printVeth.Raw != "/interface/veth/print" {
		t.Errorf("invalid print veth: %v", printVeth)
	}

	addVeth := NewAddVethInterfaceCommand(VethInterface{
		Name:    "veth1",
		Address: "172.17.0.2/24",
		Gateway: "172.17.0.1",
	})
	if addVeth.Raw != "/interface/veth/add" || addVeth.Args["address"] != "172.17.0.2/24" {
		t.Errorf("invalid add veth: %v", addVeth)
	}

	veths := ParseVethInterfaces([]map[string]string{
		{".id": "*1", "name": "veth1", "address": "172.17.0.2/24", "running": "yes", "disabled": "false"},
	})
	if len(veths) != 1 || !veths[0].Running || veths[0].Disabled {
		t.Errorf("unexpected veths parse: %+v", veths)
	}
}
