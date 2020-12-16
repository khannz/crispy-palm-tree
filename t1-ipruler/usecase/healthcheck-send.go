package usecase

import "github.com/khannz/crispy-palm-tree/lbost1a-ipRule/domain"

type HealthcheckSenderEntity struct {
	OrchestratorWorker domain.OrchestratorWorker
}

func NewHealthcheckSenderEntity(OrchestratorWorker domain.OrchestratorWorker) *HealthcheckSenderEntity {
	return &HealthcheckSenderEntity{OrchestratorWorker: OrchestratorWorker}
}

func (HealthcheckSenderEntity *HealthcheckSenderEntity) SendToHC(runtimeConfig map[string]struct{}, id string) error {
	return HealthcheckSenderEntity.OrchestratorWorker.SendRouteRuntimeConfig(runtimeConfig, id)
}
