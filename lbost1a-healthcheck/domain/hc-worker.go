package domain

// HCWorker ...
type HCWorker interface {
	// StartHealthchecksForServices([]*HCService, string) error
	NewServiceToHealtchecks(*HCService, string) error
	RemoveServiceFromHealtchecks(*HCService, string) error
	UpdateServiceAtHealtchecks(*HCService, string) (*HCService, error)
	GetServiceState(*HCService, string) (*HCService, error)
	GetServicesState(string) ([]*HCService, error)
}
