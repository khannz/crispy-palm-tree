package usecase

import (
	"errors"

	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/khannz/crispy-palm-tree/portadapter"
	"github.com/sirupsen/logrus"
)

const getNlbServicesEntity = "get-nlb-services"

// GetAllServices ...
type GetAllServices struct {
	cacheStorage      *portadapter.StorageEntity // so dirty
	locker            *domain.Locker
	gracefullShutdown *domain.GracefullShutdown
	logging           *logrus.Logger
}

// NewGetAllServices ...
func NewGetAllServices(cacheStorage *portadapter.StorageEntity,
	locker *domain.Locker,
	gracefullShutdown *domain.GracefullShutdown,
	logging *logrus.Logger) *GetAllServices {
	return &GetAllServices{
		cacheStorage:      cacheStorage,
		locker:            locker,
		gracefullShutdown: gracefullShutdown,
		logging:           logging,
	}
}

// GetAllServices ...
func (getAllServices *GetAllServices) GetAllServices(getAllServicesRequestUUID string) ([]*domain.ServiceInfo, error) {
	getAllServices.locker.Lock()
	defer getAllServices.locker.Unlock()
	getAllServices.gracefullShutdown.Lock()
	if getAllServices.gracefullShutdown.ShutdownNow {
		defer getAllServices.gracefullShutdown.Unlock()
		return nil, errors.New("program got shutdown signal, job get services cancel")
	}
	getAllServices.gracefullShutdown.UsecasesJobs++
	getAllServices.gracefullShutdown.Unlock()
	defer decreaseJobs(getAllServices.gracefullShutdown)

	return getAllServices.cacheStorage.LoadAllStorageDataToDomainModel()
}
