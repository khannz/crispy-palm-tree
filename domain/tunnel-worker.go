package domain

// TunnelMaker ...
type TunnelMaker interface {
	// EnrichApplicationServersInfo([]*ApplicationServer, string) ([]*ApplicationServer, error)
	// EnrichApplicationServerInfo(*ApplicationServer, int, string) (*ApplicationServer, error)
	CreateTunnel(*TunnelForApplicationServer, string) error
	CreateTunnels([]*TunnelForApplicationServer, string) ([]*TunnelForApplicationServer, error)
	RemoveTunnel(*TunnelForApplicationServer, string) error
	RemoveTunnels([]*TunnelForApplicationServer, string) ([]*TunnelForApplicationServer, error)
	ExecuteCommandForTunnel(string, string, string) error
}
