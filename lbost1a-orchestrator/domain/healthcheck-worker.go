package domain

// HCWorker ...
type HCWorker interface {
	StartHealthchecksForCurrentServices([]*ServiceInfo, string) error
	NewServiceToHealtchecks(*ServiceInfo, string) error
	RemoveServiceFromHealtchecks(*ServiceInfo, string) error
	UpdateServiceAtHealtchecks(*ServiceInfo, string) (*ServiceInfo, error)
	GetServiceState(*ServiceInfo, string) (*ServiceInfo, error)
	GetServicesState(string) ([]*ServiceInfo, error)
	ConnectToHealtchecks() error
	DisconnectFromHealtchecks()
}
