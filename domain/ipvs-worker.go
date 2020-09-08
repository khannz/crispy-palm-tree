package domain

// IPVSWorker ...
type IPVSWorker interface {
	CreateService(*ServiceInfo, string) error
	RemoveService(*ServiceInfo, string) error
	AddApplicationServersForService(*ServiceInfo, string) error
	RemoveApplicationServersFromService(*ServiceInfo, string) error
	Flush() error
	ReadCurrentConfig() ([]*ServiceInfo, error)
}
