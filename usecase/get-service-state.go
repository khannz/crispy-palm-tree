package usecase

import (
	"fmt"

	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/khannz/crispy-palm-tree/portadapter"
)

// GetServiceStateEntity ...
type GetServiceStateEntity struct {
	locker            *domain.Locker
	cacheStorage      *portadapter.StorageEntity // so dirty
	gracefullShutdown *domain.GracefullShutdown
}

// NewGetServiceStateEntity ...
func NewGetServiceStateEntity(locker *domain.Locker,
	cacheStorage *portadapter.StorageEntity,
	gracefullShutdown *domain.GracefullShutdown) *GetServiceStateEntity {
	return &GetServiceStateEntity{
		locker:            locker,
		cacheStorage:      cacheStorage,
		gracefullShutdown: gracefullShutdown,
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

	return getServiceStateEntity.cacheStorage.GetServiceInfo(serviceInfo, getServiceStateUUID)
}