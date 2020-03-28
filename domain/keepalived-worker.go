package domain

// KeepalivedCustomizer ...
type KeepalivedCustomizer interface {
	CustomizeKeepalived(string, string, map[string]string, map[string][]string, string) (map[string][]string, error)
	RemoveKeepalivedDConfigFile(string, string) error
	RemoveKeepalivedSymlink(string, string) error
	DetectKeepalivedConfigFiles(string, string, map[string][]string, string) (map[string][]string, error)
	ReloadKeepalived(string) error
	GetInfoAboutAllNWBServices(string) ([]ServiceInfo, error)
}
