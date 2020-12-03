package domain

// MemoryWorker ...
type MemoryWorker interface {
	AddService(*ServiceInfo) error
	GetService(string) (*ServiceInfo, error)
	RemoveService(*ServiceInfo) error
	UpdateService(*ServiceInfo) error
	NeedTunnelForApplicationServer(string) bool
}
