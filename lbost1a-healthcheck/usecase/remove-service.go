package usecase

import "github.com/khannz/crispy-palm-tree/lbost1a-healthcheck/domain"

type RemoveServiceEntity struct {
	hc     domain.HCWorker
	locker *domain.Locker
}

func NewRemoveServiceEntity(hc domain.HCWorker,
	locker *domain.Locker) *RemoveServiceEntity {
	return &RemoveServiceEntity{hc: hc, locker: locker}
}

func (removeServiceEntity *RemoveServiceEntity) RemoveService(service *domain.HCService) error {
	removeServiceEntity.locker.Lock()
	defer removeServiceEntity.locker.Unlock()
	return removeServiceEntity.hc.RemoveServiceFromHealtchecks(service)
}
