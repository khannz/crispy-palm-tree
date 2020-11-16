package usecase

import "github.com/khannz/crispy-palm-tree/lbost1a-ipvs/domain"

type HealthcheckSenderEntity struct {
	healthcheckWorker domain.HealthcheckWorker
}

func NewHealthcheckSenderEntity(HealthcheckWorker domain.HealthcheckWorker) *HealthcheckSenderEntity {
	return &HealthcheckSenderEntity{healthcheckWorker: HealthcheckWorker}
}

func (HealthcheckSenderEntity *HealthcheckSenderEntity) SendToHC(runtimeConfig map[string]map[string]uint16, id string) error {
	return HealthcheckSenderEntity.healthcheckWorker.SendIPVSRuntime(runtimeConfig, id)
}
