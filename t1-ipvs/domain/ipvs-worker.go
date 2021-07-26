package domain

// IPVSWorker ...
type IPVSWorker interface {
	NewIPVSService(id, balanceType, vip string, protocol, port uint16, routingType uint32) error
	AddIPVSApplicationServersForService(id, balanceType, vip string, protocol, port uint16, routingType uint32, applicationServers map[string]uint16) error
	RemoveIPVSService(id, vip string, protocol, port uint16) error
	RemoveIPVSApplicationServersFromService(id, balanceType, vip string, protocol, port uint16, routingType uint32, applicationServers map[string]uint16) error
	GetIPVSRuntime(id string) (map[string]map[string]uint16, error)
}
