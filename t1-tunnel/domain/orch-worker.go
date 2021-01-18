package domain

// OrchestratorWorker ...
type OrchestratorWorker interface {
	TunnelRuntimeConfig(map[string]struct{}, string) error
}
