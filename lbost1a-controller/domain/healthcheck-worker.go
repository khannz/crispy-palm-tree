package domain

// HCWorker ...
type HCWorker interface {
	StartHealthchecksForCurrentServices([]*ServiceInfo) error
	NewServiceToHealtchecks(*ServiceInfo) error
	RemoveServiceFromHealtchecks(*ServiceInfo) error
	UpdateServiceAtHealtchecks(*ServiceInfo) (*ServiceInfo, error)
	GetServiceState(*ServiceInfo) (*ServiceInfo, error)
	GetServicesState() ([]*ServiceInfo, error)
	ConnectToHealtchecks() error
	DisconnectFromHealtchecks()
}
