package domain

// OrchestratorWorker ...
type OrchestratorWorker interface {
	SendIPVSRuntime(map[string]map[string]uint16, string) error
}
