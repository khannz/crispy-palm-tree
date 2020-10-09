package domain

// IPVSWorker ...
type IPVSWorker interface {
	NewService(string, uint16, uint32, string, uint16, map[string]uint16, string) error
	RemoveService(string, uint16, uint16, string) error
	AddApplicationServersForService(string, uint16, uint32, string, uint16, map[string]uint16, string) error
	RemoveApplicationServersFromService(string, uint16, uint32, string, uint16, map[string]uint16, string) error
	IsApplicationServerInService(string, uint16, string, uint16) (bool, error)
	Flush() error
}
