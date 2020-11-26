package domain

// MemoryWorker ...
type MemoryWorker interface {
	AddService(*ServiceInfo) error
	GetService(string) (*ServiceInfo, error)
	RemoveService()
	UpdateService(*ServiceInfo) error
}
