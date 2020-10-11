package domain

// IPVSSender ...
type IPVSSender interface {
	NewIPVSService(string, uint16, uint32, string, uint16, map[string]uint16, string) error
	AddIPVSApplicationServersForService(string, uint16, uint32, string, uint16, map[string]uint16, string) error
	RemoveIPVSService(string, uint16, uint16, string) error
	RemoveIPVSApplicationServersFromService(string, uint16, uint32, string, uint16, map[string]uint16, string) error
	IsIPVSApplicationServerInService(string, uint16, map[string]uint16, string) (bool, error)
	IPVSFlush() error
}
