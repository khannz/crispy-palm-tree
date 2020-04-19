package usecase

import (
	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/khannz/crispy-palm-tree/portadapter"
	"github.com/sirupsen/logrus"
)

const getNlbServicesEntity = "get-nlb-services"

// GetAllServices ...
type GetAllServices struct {
	cacheStorage *portadapter.StorageEntity // so dirty
	locker       *domain.Locker
	logging      *logrus.Logger
}

// NewGetAllServices ...
func NewGetAllServices(cacheStorage *portadapter.StorageEntity,
	locker *domain.Locker,
	logging *logrus.Logger) *GetAllServices {
	return &GetAllServices{
		cacheStorage: cacheStorage,
		locker:       locker,
		logging:      logging,
	}
}

// GetAllServices ...
func (getAllServices *GetAllServices) GetAllServices(getAllServicesRequestUUID string) ([]domain.ServiceInfo, error) {
	getAllServices.locker.Lock()
	// TODO: check stop chan for graceful shutdown
	defer getAllServices.locker.Unlock()
	return getAllServices.cacheStorage.LoadAllStorageDataToDomainModel()
}
