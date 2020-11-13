package domain

// HealthcheckWorker ...
type HealthcheckWorker interface {
	SendRuntimeConfig(map[string]struct{}, string) error
}
