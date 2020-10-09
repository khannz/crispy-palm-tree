package usecase

import (
	"github.com/khannz/crispy-palm-tree/lbost1a-healthcheck/domain"
)

type GetServicesEntity struct {
	hc     domain.HCWorker
	locker *domain.Locker
}

func NewGetServicesEntity(hc domain.HCWorker,
	locker *domain.Locker) *GetServicesEntity {
	return &GetServicesEntity{hc: hc, locker: locker}
}

func (getServicesEntity *GetServicesEntity) GetServices() ([]*domain.HCService, error) {
	getServicesEntity.locker.Lock()
	defer getServicesEntity.locker.Unlock()
	return getServicesEntity.hc.GetServicesState()
}
