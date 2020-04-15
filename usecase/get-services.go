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
	logging      *logrus.Logger
}

// NewGetAllServices ...
func NewGetAllServices(cacheStorage *portadapter.StorageEntity,
	logging *logrus.Logger) *GetAllServices {
	return &GetAllServices{
		cacheStorage: cacheStorage,
		logging:      logging,
	}
}

// GetAllServices ...
func (getAllServices *GetAllServices) GetAllServices(getAllServicesRequestUUID string) ([]domain.ServiceInfo, error) {
	return getAllServices.cacheStorage.LoadAllStorageDataToDomainModel()
}
