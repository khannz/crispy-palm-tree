package domain

import "fmt"

// TunnelForApplicationServer ...
type TunnelForApplicationServer struct {
	ApplicationServerIP   string `json:"applicationServerIP"`
	IfcfgTunnelFile       string `json:"ifcfgTunnelFile"` // full path to ifcfg file
	RouteTunnelFile       string `json:"tunnelFile"`      // full path to route file
	SysctlConfFile        string `json:"sysctlConf"`      // full path to sysctl conf file
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
	return fmt.Sprintf("applicationServer{ApplicationServerIP:%s, IfcfgTunnelFile:%s, RouteTunnelFile:%v, SysctlConfFile:%s, TunnelName:%s,  ServicesToTunnelCount:%v}",
		tunnelForApplicationServer.ApplicationServerIP,
		tunnelForApplicationServer.IfcfgTunnelFile,
		tunnelForApplicationServer.RouteTunnelFile,
		tunnelForApplicationServer.SysctlConfFile,
		tunnelForApplicationServer.TunnelName,
		tunnelForApplicationServer.ServicesToTunnelCount)
}
