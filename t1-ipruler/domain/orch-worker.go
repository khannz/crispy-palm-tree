package domain

// OrchestratorWorker ...
type OrchestratorWorker interface {
	RouteRuntimeConfig(map[int]struct{}, string) error
}
