package usecase

import "github.com/khannz/crispy-palm-tree/lbost1a-healthcheck/domain"

type GetServiceEntity struct {
	hc     domain.HCWorker
	locker *domain.Locker
}

func NewGetServiceEntity(hc domain.HCWorker,
	locker *domain.Locker) *GetServiceEntity {
	return &GetServiceEntity{hc: hc, locker: locker}
}

func (getServiceEntity *GetServiceEntity) GetService(service *domain.HCService, id string) (*domain.HCService, error) {
	getServiceEntity.locker.Lock()
	defer getServiceEntity.locker.Unlock()
	return getServiceEntity.hc.GetServiceState(service, id)
}
