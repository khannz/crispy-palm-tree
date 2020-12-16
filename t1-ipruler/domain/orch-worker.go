package domain

// OrchestratorWorker ...
type OrchestratorWorker interface {
	SendRouteRuntimeConfig(map[int]struct{}, string) error
}
