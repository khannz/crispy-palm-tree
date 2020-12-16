package application

import (
	"time"

	"github.com/khannz/crispy-palm-tree/t1-ipruler/domain"
	"github.com/khannz/crispy-palm-tree/t1-ipruler/usecase"
	"github.com/sirupsen/logrus"
)

const sendRuntimeConfigName = "send runtime config"

// RouteFacade struct
type RouteFacade struct {
	RouteWorker        domain.RouteWorker
	OrchestratorWorker domain.OrchestratorWorker
	IDgenerator        domain.IDgenerator
	Logging            *logrus.Logger
}

// NewRouteFacade ...
func NewRouteFacade(ipRuleWorker domain.RouteWorker,
	OrchestratorWorker domain.OrchestratorWorker,
	idGenerator domain.IDgenerator,
	logging *logrus.Logger) *RouteFacade {

	return &RouteFacade{
		RouteWorker:        ipRuleWorker,
		OrchestratorWorker: OrchestratorWorker,
		IDgenerator:        idGenerator,
		Logging:            logging,
	}
}

func (ipRuleFacade *RouteFacade) AddIPRule(hcTunDestIP string, id string) error {
	newAddToRouteEntity := usecase.NewAddToRouteEntity(ipRuleFacade.RouteWorker)
	return newAddToRouteEntity.AddIPRule(hcTunDestIP, id)
}

func (ipRuleFacade *RouteFacade) RemoveIPRule(hcTunDestIP string, id string) error {
	newAddToRouteEntity := usecase.NewRemoveIPRuleEntity(ipRuleFacade.RouteWorker)
	return newAddToRouteEntity.RemoveIPRule(hcTunDestIP, id)
}

func (ipRuleFacade *RouteFacade) GetIPRuleRuntimeConfig(id string) (map[int]struct{}, error) {
	newGetRuntimeConfigEntity := usecase.NewGetRuntimeConfigEntity(ipRuleFacade.RouteWorker)
	return newGetRuntimeConfigEntity.GetIPRuleRuntimeConfig(id)
}

func (ipRuleFacade *RouteFacade) TryToSendRuntimeConfig(id string) {
	newGetRuntimeConfigEntity := usecase.NewGetRuntimeConfigEntity(ipRuleFacade.RouteWorker)
	newHealthcheckSenderEntity := usecase.NewHealthcheckSenderEntity(ipRuleFacade.OrchestratorWorker)
	for {
		currentConfig, err := newGetRuntimeConfigEntity.GetIPRuleRuntimeConfig(id)
		if err != nil {
			ipRuleFacade.Logging.WithFields(logrus.Fields{
				"entity":   sendRuntimeConfigName,
				"event id": id,
			}).Errorf("failed to get runtime config: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		if err := newHealthcheckSenderEntity.SendToHC(currentConfig, id); err != nil {
			ipRuleFacade.Logging.WithFields(logrus.Fields{
				"entity":   sendRuntimeConfigName,
				"event id": id,
			}).Debugf("failed to send runtime config: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}
		ipRuleFacade.Logging.WithFields(logrus.Fields{
			"entity":   sendRuntimeConfigName,
			"event id": id,
		}).Info("send runtime config to hc success")
		break
	}
}
