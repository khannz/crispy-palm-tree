package domain

import "fmt"

// TunnelForApplicationServer ...TODO: route table hardcoded
type TunnelForApplicationServer struct {
	ApplicationServerIP   string `json:"applicationServerIP"`
	SysctlConfFile        string `json:"sysctlConf"` // full path to sysctl conf file
	TunnelName            string `json:"tunnelName"`
	ServicesToTunnelCount int    `json:"servicesToTunnelCount"`
}

// TunnelMaker ...
type TunnelMaker interface {
	// EnrichApplicationServersInfo([]*ApplicationServer, string) ([]*ApplicationServer, error)
	// EnrichApplicationServerInfo(*ApplicationServer, int, string) (*ApplicationServer, error)
	CreateTunnel(*TunnelForApplicationServer, string) error
	CreateTunnels([]*TunnelForApplicationServer, string) ([]*TunnelForApplicationServer, error)
	RemoveTunnel(*TunnelForApplicationServer, string) error
	RemoveTunnels([]*TunnelForApplicationServer, string) ([]*TunnelForApplicationServer, error)
}

// Release stringer interface for print/log data in []*TunnelForApplicationServer
func (tunnelForApplicationServer *TunnelForApplicationServer) String() string {
	return fmt.Sprintf("applicationServer{ApplicationServerIP:%s, SysctlConfFile:%s, TunnelName:%s,  ServicesToTunnelCount:%v}",
		tunnelForApplicationServer.ApplicationServerIP,
		tunnelForApplicationServer.SysctlConfFile,
		tunnelForApplicationServer.TunnelName,
		tunnelForApplicationServer.ServicesToTunnelCount)
}
