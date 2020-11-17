package domain

// IPVSWorker ...
type IPVSWorker interface {
	NewIPVSService(string, uint16, uint32, string, uint16, string) error
	AddIPVSApplicationServersForService(string, uint16, uint32, string, uint16, map[string]uint16, string) error
	RemoveIPVSService(string, uint16, uint16, string) error
	RemoveIPVSApplicationServersFromService(string, uint16, uint32, string, uint16, map[string]uint16, string) error
	GetIPVSRuntime(string) (map[string]map[string]uint16, error)
}
