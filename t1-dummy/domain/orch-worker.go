package domain

// OrchestratorWorker ...
type OrchestratorWorker interface {
	DummyRuntimeConfig(map[string]struct{}, string) error
}
