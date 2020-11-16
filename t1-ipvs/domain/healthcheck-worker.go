package domain

// HealthcheckWorker ...
type HealthcheckWorker interface {
	SendIPVSRuntime(map[string]map[string]uint16, string) error
}
