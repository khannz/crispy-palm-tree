package usecase

import "github.com/khannz/crispy-palm-tree/t1-orch/domain"

func addTunnelRouteIpRule( // FIXME: tun make
	routeMaker domain.RouteWorker, serviceIP string,
	// FIXME: rule make
	appSrvIP string,
	id string) error {
	// if err := newService.tunnelMaker.AddTunnel(); err != nil { FIXME:
	// 	return err
	// }

	if err := routeMaker.AddRoute(serviceIP, appSrvIP, id); err != nil {
		return err
	}

	// if err := newService.ipruleMaker.AddIPRule(); err != nil { FIXME:
	// 	return err
	// }
	return nil
}

func removeRouteTunnelIpRule(routeMaker domain.RouteWorker, serviceIP string,
	// FIXME: tun rem
	// FIXME: rule rem
	appSrvIP string,
	id string) error {
	if err := routeMaker.RemoveRoute(serviceIP, appSrvIP, id); err != nil {
		return err
	}

	// if err := newService.tunnelMaker.AddTunnel(); err != nil { FIXME:
	// 	return err
	// }

	// if err := newService.ipruleMaker.AddIPRule(); err != nil { FIXME:
	// 	return err
	// }
	return nil
}
