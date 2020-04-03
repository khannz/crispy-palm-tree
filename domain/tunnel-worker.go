package domain

// TunnelMaker ...
type TunnelMaker interface {
	CreateTunnel(map[string][]string, map[string]string, string) (map[string][]string, error)
	RemoveCreatedTunnelFiles([]string, string) error
	ExecuteCommandForTunnels([]string, string, string) error
	DetectTunnels([]ApplicationServer, map[string][]string, string) (map[string][]string, error)
	RemoveApplicationServersFromTunnels([]string, map[string]string, string) error
}
