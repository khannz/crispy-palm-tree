package domain

// OrchestratorWorker ...
type OrchestratorWorker interface {
	SendTunnelRuntimeConfig(map[string]struct{}, string) error
}
