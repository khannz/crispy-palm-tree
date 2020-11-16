package domain

// HealthcheckWorker ...
type HealthcheckWorker interface {
	SendDummyRuntimeConfig(map[string]struct{}, string) error
}
