package application

import (
	"time"

	"github.com/khannz/crispy-palm-tree/lbost1a-tunnel/domain"
	"github.com/khannz/crispy-palm-tree/lbost1a-tunnel/usecase"
	"github.com/sirupsen/logrus"
)

const sendRuntimeConfigName = "send runtime config"

// TunnelFacade struct
type TunnelFacade struct {
	TunnelWorker       domain.TunnelWorker
	OrchestratorWorker domain.OrchestratorWorker
	IDgenerator        domain.IDgenerator
	Logging            *logrus.Logger
}

// NewTunnelFacade ...
func NewTunnelFacade(tunnelWorker domain.TunnelWorker,
	OrchestratorWorker domain.OrchestratorWorker,
	idGenerator domain.IDgenerator,
	logging *logrus.Logger) *TunnelFacade {

	return &TunnelFacade{
		TunnelWorker:       tunnelWorker,
		OrchestratorWorker: OrchestratorWorker,
		IDgenerator:        idGenerator,
		Logging:            logging,
	}
}

func (routeFacade *TunnelFacade) AddTunnel(hcTunDestIP string, id string) error {
	newAddToTunnelEntity := usecase.NewAddTunnelEntity(routeFacade.TunnelWorker)
	return newAddToTunnelEntity.AddTunnel(hcTunDestIP, id)
}

func (routeFacade *TunnelFacade) RemoveTunnel(hcTunDestIP string, needRemoveTunnel bool, id string) error {
	newAddToTunnelEntity := usecase.NewRemoveTunnelEntity(routeFacade.TunnelWorker)
	return newAddToTunnelEntity.RemoveTunnel(hcTunDestIP, needRemoveTunnel, id)
}

func (routeFacade *TunnelFacade) GetTunnelRuntime(id string) (map[string]struct{}, error) {
	newGetRuntimeConfigEntity := usecase.NewGetRuntimeConfigEntity(routeFacade.TunnelWorker)
	return newGetRuntimeConfigEntity.GetTunnelRuntime(id)
}

func (routeFacade *TunnelFacade) TryToSendRuntimeConfig(id string) {
	newGetRuntimeConfigEntity := usecase.NewGetRuntimeConfigEntity(routeFacade.TunnelWorker)
	newHealthcheckSenderEntity := usecase.NewHealthcheckSenderEntity(routeFacade.OrchestratorWorker)
	for {
		currentConfig, err := newGetRuntimeConfigEntity.GetTunnelRuntime(id)
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
