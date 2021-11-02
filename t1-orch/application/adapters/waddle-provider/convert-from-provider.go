package waddle_provider

import (
	"github.com/khannz/crispy-palm-tree/t1-orch/domain"
	"github.com/khannz/crispy-palm-tree/t1-orch/providers/waddle"
	"time"
)

type (
	ServiceConfigData waddle.ServiceConfigData
)

func (from ServiceConfigData) ToServiceInfoConf() (domain.ServiceInfoConf, error) {
	services := from.GetServices()
	servicesInfo := make(domain.ServiceInfoConf, len(services))
	for _, service := range services {

		_ = service.GetProtocol()
		_ = service.GetAddress()
		_ = service.GetPort()
		_ = service.GetRoutingType()
		_ = service.GetBalancingType()

		for _, anReal := range service.GetReals() {
			/*//
			var (
				err                error
				applicationServers domain.ApplicationServers
			)
			*/
			_ = anReal.GetAddress()
			_ = anReal.GetPort()

			hc := anReal.GetHealthcheck()
			quorum := int(hc.GetQuorum())
			helloTimer := time.Duration(hc.GetHelloTimer())
			responseTimer := time.Duration(hc.GetResponseTimer())
			aliveThreshold := int(hc.GetAliveThreshold())
			deadThreshold := int(hc.GetDeadThreshold())

			_ = quorum
			_ = helloTimer
			_ = responseTimer
			_ = aliveThreshold
			_ = deadThreshold

		}
	}
	return servicesInfo, nil
}
