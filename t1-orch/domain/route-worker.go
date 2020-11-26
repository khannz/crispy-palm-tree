package domain

import "fmt"

// RouteForApplicationServer ...TODO: route table hardcoded
type RouteForApplicationServer struct {
	ApplicationServerIP   string `json:"applicationServerIP"`
	SysctlConfFile        string `json:"sysctlConf"` // full path to sysctl conf file
	TunnelName            string `json:"tunnelName"`
	ServicesToTunnelCount int    `json:"servicesToTunnelCount"`
}

// RouteMaker ...
type RouteMaker interface {
	AddRoute(string, string, string) error
	RemoveRoute(string, string, bool, string) error
	GetRouteRuntimeConfig(string) ([]string, error)
}

// Release stringer interface for print/log data in []*RouteForApplicationServer
func (routeForApplicationServer *RouteForApplicationServer) String() string {
	return fmt.Sprintf("applicationServer{ApplicationServerIP:%v, SysctlConfFile:%v, TunnelName:%v,  ServicesToTunnelCount:%v}",
		routeForApplicationServer.ApplicationServerIP,
		routeForApplicationServer.SysctlConfFile,
		routeForApplicationServer.TunnelName,
		routeForApplicationServer.ServicesToTunnelCount)
}
