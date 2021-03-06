package domain

// MemoryWorker ...
type MemoryWorker interface {
	AddService(*ServiceInfo) error
	GetService(string) (*ServiceInfo, error)
	GetServices() map[string]*ServiceInfo
	RemoveService(*ServiceInfo) error
	UpdateService(*ServiceInfo) error
	AddTunnelForApplicationServer(string) bool
	RemoveTunnelForApplicationServer(string) bool
}
