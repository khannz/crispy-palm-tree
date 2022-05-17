package waddle_provider

import (
	"fmt"
	"time"

	"github.com/khannz/crispy-palm-tree/t1-orch/domain"
	"github.com/khannz/crispy-palm-tree/t1-orch/providers/waddle"
)

type (
	ServiceConfigData waddle.ServiceConfigData
)

func (from ServiceConfigData) ToServiceInfoConf() (domain.ServiceInfoConf, error) {
	services := from.GetServices()

	servicesInfo := make(domain.ServiceInfoConf, len(services))
	for _, service := range services {
		r := map[string]*domain.ApplicationServer{}
		srv := &domain.ServiceInfo{}

		// TODO: no idea what is purpose of Address
		srv.Address = fmt.Sprintf("%s:%d", service.GetAddress(), service.GetPort())

		srv.Protocol = service.GetProtocol().String()
		srv.IP = service.GetAddress()
		srv.Port = fmt.Sprintf("%d", service.GetPort())

		srv.BalanceType = service.GetBalancingType().String()
		srv.RoutingType = service.GetRoutingType().String()

		// TODO: srv.HealthcheckType = service.Get

		for _, anReal := range service.GetReals() {
			/*//
			var (
				err                error
				applicationServers domain.ApplicationServers
			)
			*/

			/*
				FIXME: until migration to protos-v2.* this logic with overwriting is acceptable
				since following values in protos-v1.* should be shared for every Real
			*/
			hc := anReal.GetHealthcheck()
			srv.HelloTimer = time.Duration(hc.GetHelloTimer()) * time.Millisecond
			srv.ResponseTimer = time.Duration(hc.GetResponseTimer()) * time.Millisecond
			srv.AliveThreshold = int(hc.AliveThreshold)
			srv.DeadThreshold = int(hc.DeadThreshold)
			srv.HealthcheckType = hc.Type
			//srv.Uri = hc.Uri // TODO: does it implemented in core logic
			//srv.ValidResponseCodes = hc.ResponseCodes // TODO: does it implemented in core logic

			res := &domain.ApplicationServer{
				Address:            fmt.Sprintf("%s:%d", anReal.GetAddress(), anReal.GetPort()),
				IP:                 anReal.GetAddress(),
				Port:               fmt.Sprintf("%d", anReal.GetPort()),
				HealthcheckAddress: fmt.Sprintf("%s:%d", hc.Address, hc.Port),
			}

			r[res.Address] = res
		}
		srv.ApplicationServers = r
		servicesInfo[srv.Address] = srv
	}
	return servicesInfo, nil
}
