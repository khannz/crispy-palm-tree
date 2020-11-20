package domain

// OrchestratorWorker ...
type OrchestratorWorker interface {
	SendDummyRuntimeConfig(map[string]struct{}, string) error
}
