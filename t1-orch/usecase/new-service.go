package usecase

import (
	"fmt"

	"github.com/khannz/crispy-palm-tree/t1-orch/domain"
	"github.com/khannz/crispy-palm-tree/t1-orch/healthcheck"
	"github.com/sirupsen/logrus"
)

const newServiceName = "new service"

// NewServiceEntity ...
type NewServiceEntity struct {
	memoryWorker     domain.MemoryWorker
	routeMaker       domain.RouteWorker
	hc               *healthcheck.HeathcheckEntity
	gracefulShutdown *domain.GracefulShutdown
	logging          *logrus.Logger
}

// NewNewServiceEntity ... // TODO: naming
func NewNewServiceEntity(memoryWorker domain.MemoryWorker,
	routeMaker domain.RouteWorker,
	hc *healthcheck.HeathcheckEntity,
	gracefulShutdown *domain.GracefulShutdown,
	logging *logrus.Logger) *NewServiceEntity {
	return &NewServiceEntity{
		memoryWorker:     memoryWorker,
		routeMaker:       routeMaker,
		hc:               hc,
		gracefulShutdown: gracefulShutdown,
		logging:          logging,
	}
}

// NewService ...
func (newService *NewServiceEntity) NewService(serviceInfo *domain.ServiceInfo,
	newServiceID string) error {
	// graceful shutdown part start
	newService.gracefulShutdown.Lock()
	if newService.gracefulShutdown.ShutdownNow {
		defer newService.gracefulShutdown.Unlock()
		return fmt.Errorf("got shutdown signal, job new service %v canceled", serviceInfo)
	}
	newService.gracefulShutdown.UsecasesJobs++
	newService.gracefulShutdown.Unlock()
	defer decreaseJobs(newService.gracefulShutdown)
	// graceful shutdown part end
	newService.logging.WithFields(logrus.Fields{
		"entity":   newServiceName,
		"event id": newServiceID,
	}).Infof("start usecase for new service: %v", serviceInfo)

	if serviceInfo.RoutingType == "tunneling" {
		for _, appSrv := range serviceInfo.ApplicationServers { // TODO: "nat not ready, only tcp at now"
			if err := newService.routeMaker.AddRoute(serviceInfo.IP, appSrv.IP, newServiceID); err != nil {
				return err
			}
		}
	}

	if err := newService.memoryWorker.AddService(serviceInfo); err != nil {
		return err
	}

	newService.logging.WithFields(logrus.Fields{
		"entity":   newServiceName,
		"event id": newServiceID,
	}).Info("update service at healtchecks")
	if err := newService.hc.NewServiceToHealtchecks(serviceInfo, newServiceID); err != nil {
		return fmt.Errorf("error when change service in healthcheck: %v", err)
	}
	newService.logging.WithFields(logrus.Fields{
		"entity":   newServiceName,
		"event id": newServiceID,
	}).Info("updated service at healtchecks")
	return nil
}