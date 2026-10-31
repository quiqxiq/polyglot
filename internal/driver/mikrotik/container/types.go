package container

// Container represents an OCI container instance in RouterOS (/container).
type Container struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Tag       string `json:"tag"`
	Status    string `json:"status"`
	Running   bool   `json:"running"`
	Image     string `json:"image"`
	RootDir   string `json:"root_dir"`
	Cmd       string `json:"cmd"`
	Interface string `json:"interface"`
	Comment   string `json:"comment"`
}

// ContainerConfig represents global container daemon settings (/container/config).
type ContainerConfig struct {
	RegistryURL string `json:"registry_url"`
	TmpDir      string `json:"tmpdir"`
}

// AppEntry represents a packaged container app from the RouterOS 7.21+ App Store (/app).
type AppEntry struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Running     bool   `json:"running"`
	Status      string `json:"status"`
	UIURL       string `json:"ui_url"`
	Interface   string `json:"interface"`
	AppSize     int64  `json:"app_size"`
	DataSize    int64  `json:"data_size"`
}
