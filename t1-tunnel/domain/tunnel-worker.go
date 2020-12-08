package domain

// TunnelWorker ...
type TunnelWorker interface {
	AddTunnel(string, string) error
	RemoveTunnel(string, bool, string) error
	GetTunnelRuntime(string) (map[string]struct{}, error)
}
