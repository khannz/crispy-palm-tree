package usecase

import "github.com/khannz/crispy-palm-tree/t1-orch/domain"

func addTunnelRouteIpRule(tunnelMaker domain.TunnelWorker,
	routeMaker domain.RouteWorker,
	ipRuleWorker domain.IpRuleWorker,
	serviceIP string,
	appSrvIP string,
	id string) error {
	if err := tunnelMaker.AddTunnel(appSrvIP, id); err != nil {
		return err
	}
	// TODO: appSrvIP may be not dest for tun. need rework that (not only at T1)
	if err := routeMaker.AddRoute(serviceIP, appSrvIP, id); err != nil {
		return err
	}

	if err := ipRuleWorker.AddIPRule(appSrvIP, id); err != nil {
		return err
	}
	return nil
}

func removeRouteTunnelIpRule(routeMaker domain.RouteWorker,
	tunnelMaker domain.TunnelWorker,
	ipRuleWorker domain.IpRuleWorker,
	serviceIP string,
	appSrvIP string,
	id string) error {
	if err := routeMaker.RemoveRoute(serviceIP, appSrvIP, id); err != nil {
		return err
	}

	if err := tunnelMaker.RemoveTunnel(appSrvIP, id); err != nil {
		return err
	}

	if err := ipRuleWorker.RemoveIPRule(appSrvIP, id); err != nil {
		return err
	}
	return nil
}
