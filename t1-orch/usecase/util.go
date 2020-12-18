package usecase

import "github.com/khannz/crispy-palm-tree/t1-orch/domain"

func addTunnelRouteIpRule(tunnelMaker domain.TunnelWorker,
	routeMaker domain.RouteWorker, serviceIP string,
	// FIXME: rule make
	appSrvIP string,
	id string) error {
	if err := tunnelMaker.AddTunnel(appSrvIP, id); err != nil {
		return err
	}
	// FIXME: appSrvIP may be not dest fro tun. take it from HC address
	if err := routeMaker.AddRoute(serviceIP, appSrvIP, id); err != nil {
		return err
	}

	// if err := newService.ipruleMaker.AddIPRule(); err != nil { FIXME:
	// 	return err
	// }
	return nil
}

func removeRouteTunnelIpRule(routeMaker domain.RouteWorker,
	tunnelMaker domain.TunnelWorker,
	// FIXME: rule rem
	serviceIP string,
	appSrvIP string,
	id string) error {
	if err := routeMaker.RemoveRoute(serviceIP, appSrvIP, id); err != nil {
		return err
	}

	if err := tunnelMaker.RemoveTunnel(appSrvIP, id); err != nil {
		return err
	}

	// if err := newService.ipruleMaker.AddIPRule(); err != nil { FIXME:
	// 	return err
	// }
	return nil
}
