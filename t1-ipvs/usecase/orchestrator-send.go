package usecase

import "github.com/khannz/crispy-palm-tree/lbost1a-ipvs/domain"

type OrchestratorSenderEntity struct {
	orchestratorWorker domain.OrchestratorWorker
}

func NewOrchestratorSenderEntity(OrchestratorWorker domain.OrchestratorWorker) *OrchestratorSenderEntity {
	return &OrchestratorSenderEntity{orchestratorWorker: OrchestratorWorker}
}

func (orchestratorSenderEntity *OrchestratorSenderEntity) SendToOrch(runtimeConfig map[string]map[string]uint16, id string) error {
	return orchestratorSenderEntity.orchestratorWorker.SendIPVSRuntime(runtimeConfig, id)
}
