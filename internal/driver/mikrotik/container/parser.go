package container

import (
	"strconv"
)

// ParseContainers converts reply rows into Container slice.
func ParseContainers(rows []map[string]string) []Container {
	out := make([]Container, 0, len(rows))
	for _, r := range rows {
		out = append(out, Container{
			ID:        r[".id"],
			Name:      r["name"],
			Tag:       r["tag"],
			Status:    r["status"],
			Running:   r["running"] == "true",
			Image:     r["image"],
			RootDir:   r["root-dir"],
			Cmd:       r["cmd"],
			Interface: r["interface"],
			Comment:   r["comment"],
		})
	}
	return out
}

// ParseContainerConfig converts reply rows into ContainerConfig.
func ParseContainerConfig(rows []map[string]string) *ContainerConfig {
	if len(rows) == 0 {
		return nil
	}
	r := rows[0]
	return &ContainerConfig{
		RegistryURL: r["registry-url"],
		TmpDir:      r["tmpdir"],
	}
}

// ParseApps converts reply rows into AppEntry slice.
func ParseApps(rows []map[string]string) []AppEntry {
	out := make([]AppEntry, 0, len(rows))
	for _, r := range rows {
		appSz, _ := strconv.ParseInt(r["app-size"], 10, 64)
		dataSz, _ := strconv.ParseInt(r["data-size"], 10, 64)
		out = append(out, AppEntry{
			ID:          r[".id"],
			Name:        r["name"],
			Category:    r["category"],
			Description: r["description"],
			Running:     r["running"] == "true",
			Status:      r["status"],
			UIURL:       r["ui-url"],
			Interface:   r["interface"],
			AppSize:     appSz,
			DataSize:    dataSz,
		})
	}
	return out
}
