package usecase

import "github.com/khannz/crispy-palm-tree/lbost1a-healthcheck/domain"

type UpdateServiceEntity struct {
	hc     domain.HCWorker
	locker *domain.Locker
}

func NewUpdateServiceEntity(hc domain.HCWorker,
	locker *domain.Locker) *UpdateServiceEntity {
	return &UpdateServiceEntity{hc: hc, locker: locker}
}

func (updateServiceEntity *UpdateServiceEntity) UpdateService(service *domain.HCService) (*domain.HCService, error) {
	updateServiceEntity.locker.Lock()
	defer updateServiceEntity.locker.Unlock()
	return updateServiceEntity.hc.UpdateServiceAtHealtchecks(service)
}
