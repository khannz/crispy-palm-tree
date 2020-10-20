package usecase

import "github.com/khannz/crispy-palm-tree/lbost1a-healthcheck/domain"

type NewServiceEntity struct {
	hc     domain.HCWorker
	locker *domain.Locker
}

func NewNewServiceEntity(hc domain.HCWorker,
	locker *domain.Locker) *NewServiceEntity {
	return &NewServiceEntity{hc: hc, locker: locker}
}

func (newServiceEntity *NewServiceEntity) NewService(service *domain.HCService, id string) error {
	newServiceEntity.locker.Lock()
	defer newServiceEntity.locker.Unlock()
	return newServiceEntity.hc.NewServiceToHealtchecks(service, id)
}
