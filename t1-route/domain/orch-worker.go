package domain

// OrchestratorWorker ...
type OrchestratorWorker interface {
	SendRouteRuntimeConfig([]string, string) error
}
