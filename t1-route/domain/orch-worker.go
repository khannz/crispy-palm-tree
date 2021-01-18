package domain

// OrchestratorWorker ...
type OrchestratorWorker interface {
	RouteRuntimeConfig([]string, string) error
}
