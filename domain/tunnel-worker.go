package domain

// TunnelMaker ...
type TunnelMaker interface {
	CreateTunnel(map[string][]string, map[string]string, string) (map[string][]string, error)
	RemoveCreatedTunnelFiles([]string, string) error
	ExecuteCommandForTunnels([]string, string, string) error
	DetectTunnels(map[string]string, map[string][]string, string) (map[string][]string, error)
}
