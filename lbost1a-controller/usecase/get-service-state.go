package usecase

import (
	"fmt"

	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/sirupsen/logrus"
)

const getServiceStateName = "get-service-state"

// GetServiceStateEntity ...
type GetServiceStateEntity struct {
	locker           *domain.Locker
	hc               domain.HCWorker
	gracefulShutdown *domain.GracefulShutdown
	logging          *logrus.Logger
}

// NewGetServiceStateEntity ...
func NewGetServiceStateEntity(locker *domain.Locker,
	hc domain.HCWorker,
	gracefulShutdown *domain.GracefulShutdown,
	logging *logrus.Logger) *GetServiceStateEntity {
	return &GetServiceStateEntity{
		locker:           locker,
		hc:               hc,
		gracefulShutdown: gracefulShutdown,
		logging:          logging,
	}
}

// GetServiceState ...
func (getServiceStateEntity *GetServiceStateEntity) GetServiceState(serviceInfo *domain.ServiceInfo,
	getServiceStateID string) (*domain.ServiceInfo, error) {
	// graceful shutdown part start
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
	// graceful shutdown part end
	logStartUsecase(getServiceStateName, "get service state", getServiceStateID, serviceInfo, getServiceStateEntity.logging)
	return getServiceStateEntity.hc.GetServiceState(serviceInfo)
}
