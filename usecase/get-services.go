package usecase

import (
	"errors"

	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/sirupsen/logrus"
)

const getAllServicesName = "get-all-services"

// GetAllServices ...
type GetAllServices struct {
	locker           *domain.Locker
	cacheStorage     domain.StorageActions
	gracefulShutdown *domain.GracefulShutdown
	logging          *logrus.Logger
}

// NewGetAllServices ...
func NewGetAllServices(locker *domain.Locker,
	cacheStorage domain.StorageActions,
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

	// graceful shutdown part start
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
	// graceful shutdown part end
	logStartUsecase(getAllServicesName, "get all services", getAllServicesRequestUUID, nil, getAllServices.logging)
	return getAllServices.cacheStorage.LoadAllStorageDataToDomainModels()
}
