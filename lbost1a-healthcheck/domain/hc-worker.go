package domain

// HCWorker ...
type HCWorker interface {
	StartHealthchecksForServices([]*HCService) error
	NewServiceToHealtchecks(*HCService) error
	RemoveServiceFromHealtchecks(*HCService) error
	UpdateServiceAtHealtchecks(*HCService) (*HCService, error)
	GetServiceState(*HCService) (*HCService, error)
	GetServicesState() ([]*HCService, error)
}
