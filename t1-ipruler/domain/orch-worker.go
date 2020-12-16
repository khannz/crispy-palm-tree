package domain

// OrchestratorWorker ...
type OrchestratorWorker interface {
	SendRouteRuntimeConfig(map[string]struct{}, string) error
}
