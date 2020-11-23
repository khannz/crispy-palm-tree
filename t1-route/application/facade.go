package application

import (
	"time"

	"github.com/khannz/crispy-palm-tree/lbost1a-route/domain"
	"github.com/khannz/crispy-palm-tree/lbost1a-route/usecase"
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
func NewRouteFacade(routeWorker domain.RouteWorker,
	OrchestratorWorker domain.OrchestratorWorker,
	idGenerator domain.IDgenerator,
	logging *logrus.Logger) *RouteFacade {

	return &RouteFacade{
		RouteWorker:        routeWorker,
		OrchestratorWorker: OrchestratorWorker,
		IDgenerator:        idGenerator,
		Logging:            logging,
	}
}

func (routeFacade *RouteFacade) AddRoute(hcDestIP, hcTunDestIP string, id string) error {
	newAddToRouteEntity := usecase.NewAddToRouteEntity(routeFacade.RouteWorker)
	return newAddToRouteEntity.AddRoute(hcDestIP, hcTunDestIP, id)
}

func (routeFacade *RouteFacade) RemoveRoute(hcDestIP, hcTunDestIP string, id string) error {
	newAddToRouteEntity := usecase.NewRemoveRouteEntity(routeFacade.RouteWorker)
	return newAddToRouteEntity.RemoveRoute(hcDestIP, hcTunDestIP, id)
}

func (routeFacade *RouteFacade) GetRouteRuntimeConfig(id string) ([]string, error) {
	newGetRuntimeConfigEntity := usecase.NewGetRuntimeConfigEntity(routeFacade.RouteWorker)
	return newGetRuntimeConfigEntity.GetRouteRuntimeConfig(id)
}

func (routeFacade *RouteFacade) TryToSendRuntimeConfig(id string) {
	newGetRuntimeConfigEntity := usecase.NewGetRuntimeConfigEntity(routeFacade.RouteWorker)
	newHealthcheckSenderEntity := usecase.NewHealthcheckSenderEntity(routeFacade.OrchestratorWorker)
	for {
		currentConfig, err := newGetRuntimeConfigEntity.GetRouteRuntimeConfig(id)
		if err != nil {
			routeFacade.Logging.WithFields(logrus.Fields{
				"entity":   sendRuntimeConfigName,
				"event id": id,
			}).Errorf("failed to get runtime config: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		if err := newHealthcheckSenderEntity.SendToHC(currentConfig, id); err != nil {
			routeFacade.Logging.WithFields(logrus.Fields{
				"entity":   sendRuntimeConfigName,
				"event id": id,
			}).Debugf("failed to send runtime config: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}
		routeFacade.Logging.WithFields(logrus.Fields{
			"entity":   sendRuntimeConfigName,
			"event id": id,
		}).Info("send runtime config to hc success")
		break
	}
}
