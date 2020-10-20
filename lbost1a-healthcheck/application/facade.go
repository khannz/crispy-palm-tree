package application

import (
	domain "github.com/khannz/crispy-palm-tree/lbost1a-healthcheck/domain"
	"github.com/khannz/crispy-palm-tree/lbost1a-healthcheck/portadapter"
	"github.com/khannz/crispy-palm-tree/lbost1a-healthcheck/usecase"
	"github.com/sirupsen/logrus"
)

// HCFacade struct
type HCFacade struct {
	Locker           *domain.Locker
	CacheStorage     *portadapter.StorageEntity
	HeathcheckEntity domain.HCWorker
	IDgenerator      domain.IDgenerator
	Logging          *logrus.Logger
}

// NewHCFacade ...
func NewHCFacade(locker *domain.Locker,
	cacheStorage *portadapter.StorageEntity,
	hc domain.HCWorker,
	idGenerator domain.IDgenerator,
	logging *logrus.Logger) *HCFacade {

	return &HCFacade{
		Locker:           locker,
		CacheStorage:     cacheStorage,
		HeathcheckEntity: hc,
		IDgenerator:      idGenerator,
		Logging:          logging,
	}
}

func (hcFacade *HCFacade) HCGetService(hcService *domain.HCService, id string) (*domain.HCService, error) {
	getServiceEntity := usecase.NewGetServiceEntity(hcFacade.HeathcheckEntity, hcFacade.Locker)
	return getServiceEntity.GetService(hcService, id)
}

func (hcFacade *HCFacade) HCGetServices(id string) ([]*domain.HCService, error) {
	getServicesEntity := usecase.NewGetServicesEntity(hcFacade.HeathcheckEntity, hcFacade.Locker)
	return getServicesEntity.GetServices(id)
}

func (hcFacade *HCFacade) HCNewService(hcService *domain.HCService, id string) error {
	newServiceEntity := usecase.NewNewServiceEntity(hcFacade.HeathcheckEntity, hcFacade.Locker)
	return newServiceEntity.NewService(hcService, id)
}

func (hcFacade *HCFacade) HCUpdateService(hcService *domain.HCService, id string) (*domain.HCService, error) {
	updateServiceEntity := usecase.NewUpdateServiceEntity(hcFacade.HeathcheckEntity, hcFacade.Locker)
	return updateServiceEntity.UpdateService(hcService, id)
}

func (hcFacade *HCFacade) HCRemoveService(hcService *domain.HCService, id string) error {
	removeServiceEntity := usecase.NewRemoveServiceEntity(hcFacade.HeathcheckEntity, hcFacade.Locker)
	return removeServiceEntity.RemoveService(hcService, id)
}
