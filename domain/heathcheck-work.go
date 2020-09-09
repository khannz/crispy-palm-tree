package domain

// HeathcheckWorker ...
type HeathcheckWorker interface {
	StartGracefulShutdownControlForHealthchecks()
	StartHealthchecksForCurrentServices() error
	NewServiceToHealtchecks(*ServiceInfo)
	RemoveServiceFromHealtchecks(*ServiceInfo)
	UpdateServiceAtHealtchecks(*ServiceInfo) error
	CheckApplicationServersInService(*ServiceInfo)
	IsMockMode() bool
}
