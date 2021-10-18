package waddle_provider

import (
	"github.com/khannz/crispy-palm-tree/t1-orch/domain"
	"github.com/khannz/crispy-palm-tree/t1-orch/providers/waddle"
)

type (
	ServiceConfigData waddle.ServiceConfigData
)

func (from ServiceConfigData) ToServiceInfoConf() (domain.ServiceInfoConf, error) {
	services := from.GetServices()
	servicesInfo := make(domain.ServiceInfoConf, len(services))
	for _, item := range services {
		_ = item
	}
	return servicesInfo, nil
}
