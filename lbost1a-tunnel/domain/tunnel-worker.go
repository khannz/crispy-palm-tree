package domain

import "fmt"

// TunnelForApplicationServer ...TODO: route table hardcoded
type TunnelForApplicationServer struct {
	ApplicationServerIP   string `json:"applicationServerIP"`
	SysctlConfFile        string `json:"sysctlConf"` // full path to sysctl conf file
	TunnelName            string `json:"tunnelName"`
	ServicesToTunnelCount int    `json:"servicesToTunnelCount"`
}

// TunnelWorker ...
type TunnelWorker interface {
	CreateTunnels([]*TunnelForApplicationServer, string) error
	RemoveTunnels([]*TunnelForApplicationServer, string) error
	RemoveAllTunnels([]*TunnelForApplicationServer, string) error
}

// Release stringer interface for print/log data in []*TunnelForApplicationServer
func (tunnelForApplicationServer *TunnelForApplicationServer) String() string {
	return fmt.Sprintf("applicationServer{ApplicationServerIP:%v, SysctlConfFile:%v, TunnelName:%v,  ServicesToTunnelCount:%v}",
		tunnelForApplicationServer.ApplicationServerIP,
		tunnelForApplicationServer.SysctlConfFile,
		tunnelForApplicationServer.TunnelName,
		tunnelForApplicationServer.ServicesToTunnelCount)
}
