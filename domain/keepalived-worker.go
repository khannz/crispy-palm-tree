package domain

// KeepalivedCustomizer ...
type KeepalivedCustomizer interface {
	CustomizeKeepalived(string, string, map[string]string, map[string][]string, string) (map[string][]string, error)
	// RemoveRowFromKeepalivedConfigFile(string, string) error // TODO: remove that
	RemoveKeepalivedDConfigFile(string, string) error
	RemoveKeepalivedSymlink(string, string) error
	DetectKeepalivedConfigFiles(string, string, map[string][]string, string) (map[string][]string, error)
	ReloadKeepalived(string) error
}
