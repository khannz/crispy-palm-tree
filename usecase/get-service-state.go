package usecase

import (
	"fmt"

	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/khannz/crispy-palm-tree/portadapter"
	"github.com/sirupsen/logrus"
)

const getServiceStateName = "get-service-state"

// GetServiceStateEntity ...
type GetServiceStateEntity struct {
	locker           *domain.Locker
	cacheStorage     *portadapter.StorageEntity // so dirty
	gracefulShutdown *domain.GracefulShutdown
	logging          *logrus.Logger
}

// NewGetServiceStateEntity ...
func NewGetServiceStateEntity(locker *domain.Locker,
	cacheStorage *portadapter.StorageEntity,
	gracefulShutdown *domain.GracefulShutdown,
	logging *logrus.Logger) *GetServiceStateEntity {
	return &GetServiceStateEntity{
		locker:           locker,
		cacheStorage:     cacheStorage,
		gracefulShutdown: gracefulShutdown,
		logging:          logging,
	}
}

// GetServiceState ...
func (getServiceStateEntity *GetServiceStateEntity) GetServiceState(serviceInfo *domain.ServiceInfo,
	getServiceStateUUID string) (*domain.ServiceInfo, error) {
	// gracefull shutdown part start
	getServiceStateEntity.locker.Lock()
	defer getServiceStateEntity.locker.Unlock()
	getServiceStateEntity.gracefulShutdown.Lock()
	if getServiceStateEntity.gracefulShutdown.ShutdownNow {
		defer getServiceStateEntity.gracefulShutdown.Unlock()
		return nil, fmt.Errorf("program got shutdown signal, job get service state %v cancel", serviceInfo)
	}
	getServiceStateEntity.gracefulShutdown.UsecasesJobs++
	getServiceStateEntity.gracefulShutdown.Unlock()
	defer decreaseJobs(getServiceStateEntity.gracefulShutdown)
	// gracefull shutdown part end
	logStartUsecase(getServiceStateName, "get service state", getServiceStateUUID, serviceInfo, getServiceStateEntity.logging)
	return getServiceStateEntity.cacheStorage.GetServiceInfo(serviceInfo, getServiceStateUUID)
}
