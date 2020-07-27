package usecase

import (
	"errors"

	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/khannz/crispy-palm-tree/portadapter"
	"github.com/sirupsen/logrus"
)

const getAllServicesName = "get-all-services"

// GetAllServices ...
type GetAllServices struct {
	cacheStorage     *portadapter.StorageEntity // so dirty
	locker           *domain.Locker
	gracefulShutdown *domain.GracefulShutdown
	logging          *logrus.Logger
}

// NewGetAllServices ...
func NewGetAllServices(cacheStorage *portadapter.StorageEntity,
	locker *domain.Locker,
	gracefulShutdown *domain.GracefulShutdown,
	logging *logrus.Logger) *GetAllServices {
	return &GetAllServices{
		cacheStorage:     cacheStorage,
		locker:           locker,
		gracefulShutdown: gracefulShutdown,
		logging:          logging,
	}
}

// GetAllServices ...
func (getAllServices *GetAllServices) GetAllServices(getAllServicesRequestUUID string) ([]*domain.ServiceInfo, error) {

	// gracefull shutdown part start
	getAllServices.locker.Lock()
	defer getAllServices.locker.Unlock()
	getAllServices.gracefulShutdown.Lock()
	if getAllServices.gracefulShutdown.ShutdownNow {
		defer getAllServices.gracefulShutdown.Unlock()
		return nil, errors.New("program got shutdown signal, job get services cancel")
	}
	getAllServices.gracefulShutdown.UsecasesJobs++
	getAllServices.gracefulShutdown.Unlock()
	defer decreaseJobs(getAllServices.gracefulShutdown)
	// gracefull shutdown part end
	logStartUsecase(getAllServicesName, "get all services", getAllServicesRequestUUID, nil, getAllServices.logging)
	return getAllServices.cacheStorage.LoadAllStorageDataToDomainModel()
}
