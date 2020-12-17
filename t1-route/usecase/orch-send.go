package usecase

import "github.com/khannz/crispy-palm-tree/t1-route/domain"

type OrchSenderEntity struct {
	OrchestratorWorker domain.OrchestratorWorker
}

func NewOrchSenderEntity(OrchestratorWorker domain.OrchestratorWorker) *OrchSenderEntity {
	return &OrchSenderEntity{OrchestratorWorker: OrchestratorWorker}
}

func (orchSenderEntity *OrchSenderEntity) SendToOrch(runtimeConfig []string, id string) error {
	return orchSenderEntity.OrchestratorWorker.SendRouteRuntimeConfig(runtimeConfig, id)
}
