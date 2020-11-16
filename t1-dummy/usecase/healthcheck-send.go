package usecase

import "github.com/khannz/crispy-palm-tree/lbost1a-dummy/domain"

type HealthcheckSenderEntity struct {
	healthcheckWorker domain.HealthcheckWorker
}

func NewHealthcheckSenderEntity(HealthcheckWorker domain.HealthcheckWorker) *HealthcheckSenderEntity {
	return &HealthcheckSenderEntity{healthcheckWorker: HealthcheckWorker}
}

func (HealthcheckSenderEntity *HealthcheckSenderEntity) SendToHC(runtimeConfig map[string]struct{}, id string) error {
	return HealthcheckSenderEntity.healthcheckWorker.SendRuntimeConfig(runtimeConfig, id)
}
