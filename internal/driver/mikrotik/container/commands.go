package container

import (
	"github.com/quixiq/polyglot/internal/domain/command"
	"github.com/quixiq/polyglot/internal/driver/mikrotik/internal/rosutil"
)

// NewPrintContainersCommand queries installed containers with strict proplist.
func NewPrintContainersCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, ".id", "name", "tag", "status", "running", "image", "root-dir", "cmd", "interface", "comment")
	return command.Command{
		Raw:  "/container/print",
		Args: args,
	}
}

// NewStartContainerCommand starts a container by ID.
func NewStartContainerCommand(id string) command.Command {
	return command.Command{
		Raw: "/container/start",
		Args: map[string]string{
			".id": id,
		},
	}
}

// NewStopContainerCommand stops a running container by ID.
func NewStopContainerCommand(id string) command.Command {
	return command.Command{
		Raw: "/container/stop",
		Args: map[string]string{
			".id": id,
		},
	}
}

// NewRemoveContainerCommand deletes a container by ID.
func NewRemoveContainerCommand(id string) command.Command {
	return command.Command{
		Raw: "/container/remove",
		Args: map[string]string{
			".id": id,
		},
	}
}

// NewPrintContainerConfigCommand queries global container settings.
func NewPrintContainerConfigCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, "registry-url", "tmpdir")
	return command.Command{
		Raw:  "/container/config/print",
		Args: args,
	}
}

// NewSetContainerConfigCommand sets global container registry and tmpdir.
func NewSetContainerConfigCommand(registryURL, tmpDir string) command.Command {
	return command.Command{
		Raw: "/container/config/set",
		Args: map[string]string{
			"registry-url": registryURL,
			"tmpdir":       tmpDir,
		},
	}
}

// NewPrintAppsCommand queries the RouterOS App Store (/app).
func NewPrintAppsCommand() command.Command {
	args := make(map[string]string)
	rosutil.SetProplist(args, ".id", "name", "category", "description", "running", "status", "ui-url", "interface", "app-size", "data-size")
	return command.Command{
		Raw:  "/app/print",
		Args: args,
	}
}

// NewEnableAppCommand installs or starts an app in the App Store.
func NewEnableAppCommand(id string) command.Command {
	return command.Command{
		Raw: "/app/enable",
		Args: map[string]string{
			"numbers": id,
		},
	}
}

// NewDisableAppCommand stops an app in the App Store.
func NewDisableAppCommand(id string) command.Command {
	return command.Command{
		Raw: "/app/disable",
		Args: map[string]string{
			"numbers": id,
		},
	}
}

// NewCleanupAppCommand removes an app and purges its stored data.
func NewCleanupAppCommand(id string) command.Command {
	return command.Command{
		Raw: "/app/cleanup",
		Args: map[string]string{
			"numbers": id,
		},
	}
}
