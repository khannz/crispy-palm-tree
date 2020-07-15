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
	locker            *domain.Locker
	cacheStorage      *portadapter.StorageEntity // so dirty
	gracefullShutdown *domain.GracefullShutdown
	logging           *logrus.Logger
}

// NewGetServiceStateEntity ...
func NewGetServiceStateEntity(locker *domain.Locker,
	cacheStorage *portadapter.StorageEntity,
	gracefullShutdown *domain.GracefullShutdown,
	logging *logrus.Logger) *GetServiceStateEntity {
	return &GetServiceStateEntity{
		locker:            locker,
		cacheStorage:      cacheStorage,
		gracefullShutdown: gracefullShutdown,
		logging:           logging,
	}
}

// GetServiceState ...
func (getServiceStateEntity *GetServiceStateEntity) GetServiceState(serviceInfo *domain.ServiceInfo,
	getServiceStateUUID string) (*domain.ServiceInfo, error) {
	// gracefull shutdown part start
	getServiceStateEntity.locker.Lock()
	defer getServiceStateEntity.locker.Unlock()
	getServiceStateEntity.gracefullShutdown.Lock()
	if getServiceStateEntity.gracefullShutdown.ShutdownNow {
		defer getServiceStateEntity.gracefullShutdown.Unlock()
		return nil, fmt.Errorf("program got shutdown signal, job get service state %v cancel", serviceInfo)
	}
	getServiceStateEntity.gracefullShutdown.UsecasesJobs++
	getServiceStateEntity.gracefullShutdown.Unlock()
	defer decreaseJobs(getServiceStateEntity.gracefullShutdown)
	// gracefull shutdown part end
	logStartUsecase(getServiceStateName, "get service state", getServiceStateUUID, serviceInfo, getServiceStateEntity.logging)
	return getServiceStateEntity.cacheStorage.GetServiceInfo(serviceInfo, getServiceStateUUID)
}
