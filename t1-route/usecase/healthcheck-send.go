package usecase

import "github.com/khannz/crispy-palm-tree/t1-route/domain"

type HealthcheckSenderEntity struct {
	OrchestratorWorker domain.OrchestratorWorker
}

func NewHealthcheckSenderEntity(OrchestratorWorker domain.OrchestratorWorker) *HealthcheckSenderEntity {
	return &HealthcheckSenderEntity{OrchestratorWorker: OrchestratorWorker}
}

func (HealthcheckSenderEntity *HealthcheckSenderEntity) SendToHC(runtimeConfig []string, id string) error {
	return HealthcheckSenderEntity.OrchestratorWorker.SendRouteRuntimeConfig(runtimeConfig, id)
}
