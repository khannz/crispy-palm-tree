package usecase

import (
	"fmt"

	"github.com/khannz/crispy-palm-tree/t1-orch/domain"
	"github.com/khannz/crispy-palm-tree/t1-orch/healthcheck"
	"github.com/sirupsen/logrus"
)

const removeServiceName = "new service"

// RemoveServiceEntity ...
type RemoveServiceEntity struct {
	memoryWorker     domain.MemoryWorker
	routeMaker       domain.RouteWorker
	hc               *healthcheck.HeathcheckEntity
	gracefulShutdown *domain.GracefulShutdown
	logging          *logrus.Logger
}

// NewRemoveServiceEntity ... // TODO: naming
func NewRemoveServiceEntity(memoryWorker domain.MemoryWorker,
	routeMaker domain.RouteWorker,
	hc *healthcheck.HeathcheckEntity,
	gracefulShutdown *domain.GracefulShutdown,
	logging *logrus.Logger) *RemoveServiceEntity {
	return &RemoveServiceEntity{
		memoryWorker:     memoryWorker,
		routeMaker:       routeMaker,
		hc:               hc,
		gracefulShutdown: gracefulShutdown,
		logging:          logging,
	}
}

// RemoveService ...
func (removeService *RemoveServiceEntity) RemoveService(serviceInfo *domain.ServiceInfo,
	removeServiceID string) error {
	// graceful shutdown part start
	removeService.gracefulShutdown.Lock()
	if removeService.gracefulShutdown.ShutdownNow {
		defer removeService.gracefulShutdown.Unlock()
		return fmt.Errorf("got shutdown signal, job new service %v canceled", serviceInfo)
	}
	removeService.gracefulShutdown.UsecasesJobs++
	removeService.gracefulShutdown.Unlock()
	defer decreaseJobs(removeService.gracefulShutdown)
	// graceful shutdown part end
	removeService.logging.WithFields(logrus.Fields{
		"entity":   removeServiceName,
		"event id": removeServiceID,
	}).Infof("start usecase for remove service: %v", serviceInfo)

	if serviceInfo.RoutingType == "tunneling" {
		for _, appSrv := range serviceInfo.ApplicationServers { // TODO: "nat not ready, only tcp at now"
			tunnelStillNeeded := removeService.memoryWorker.NeedTunnelForApplicationServer(appSrv.IP)
			if err := removeService.routeMaker.RemoveRoute(serviceInfo.IP, appSrv.IP, !tunnelStillNeeded, removeServiceID); err != nil { // FIXME: never remove tunnels
				return err
			}
		}
	}

	removeService.logging.WithFields(logrus.Fields{
		"entity":   removeServiceName,
		"event id": removeServiceID,
	}).Info("remove service from healtchecks")
	if err := removeService.hc.RemoveServiceFromHealtchecks(serviceInfo, removeServiceID); err != nil {
		return fmt.Errorf("error when change service in healthcheck: %v", err)
	}
	removeService.logging.WithFields(logrus.Fields{
		"entity":   removeServiceName,
		"event id": removeServiceID,
	}).Info("remove service from healtchecks")
	return nil
}
