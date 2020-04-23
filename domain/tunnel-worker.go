package domain

// TunnelMaker ...
type TunnelMaker interface {
	EnrichApplicationServersInfo([]*ApplicationServer, string) ([]*ApplicationServer, error)
	EnrichApplicationServerInfo(*ApplicationServer, string) (*ApplicationServer, error)
	CreateTunnel(*ApplicationServer, string) error
	CreateTunnels([]*ApplicationServer, string) error
	RemoveTunnel(*ApplicationServer, string) error
	RemoveTunnels([]*ApplicationServer, string) error
	ExecuteCommandForTunnel(string, string, string) error
}
