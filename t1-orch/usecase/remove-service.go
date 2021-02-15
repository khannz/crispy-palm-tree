package usecase

import (
	"fmt"

	"github.com/khannz/crispy-palm-tree/t1-orch/domain"
	"github.com/khannz/crispy-palm-tree/t1-orch/healthcheck"
	"github.com/sirupsen/logrus"
)

const removeServiceName = "remove service"

// RemoveServiceEntity ...
type RemoveServiceEntity struct {
	memoryWorker     domain.MemoryWorker
	tunnelMaker      domain.TunnelWorker
	routeMaker       domain.RouteWorker
	ipRuleWorker     domain.IpRuleWorker
	hc               *healthcheck.HealthcheckEntity
	gracefulShutdown *domain.GracefulShutdown
	logging          *logrus.Logger
}

// NewRemoveServiceEntity ... // TODO: naming
func NewRemoveServiceEntity(memoryWorker domain.MemoryWorker,
	tunnelMaker domain.TunnelWorker,
	routeMaker domain.RouteWorker,
	ipRuleWorker domain.IpRuleWorker,
	hc *healthcheck.HealthcheckEntity,
	gracefulShutdown *domain.GracefulShutdown,
	logging *logrus.Logger) *RemoveServiceEntity {
	return &RemoveServiceEntity{
		memoryWorker:     memoryWorker,
		tunnelMaker:      tunnelMaker,
		routeMaker:       routeMaker,
		ipRuleWorker:     ipRuleWorker,
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
		for _, appSrv := range serviceInfo.ApplicationServers {
			needRemoveTunnel := removeService.memoryWorker.RemoveTunnelForApplicationServer(appSrv.IP)
			if needRemoveTunnel {
				if err := removeRouteTunnelIpRule(removeService.routeMaker,
					removeService.tunnelMaker,
					removeService.ipRuleWorker,
					serviceInfo.IP,
					appSrv.IP,
					removeServiceID); err != nil {
					return err
				}
			}
		}
	}

	removeService.logging.WithFields(logrus.Fields{
		"entity":   removeServiceName,
		"event id": removeServiceID,
	}).Info("remove service from healthchecks")
	if err := removeService.hc.RemoveServiceFromHealthchecks(serviceInfo, removeServiceID); err != nil {
		return fmt.Errorf("error when remove service in healthcheck: %v", err)
	}

	if err := removeService.memoryWorker.RemoveService(serviceInfo); err != nil {
		return fmt.Errorf("error when remove service in memory: %v", err)
	}

	return nil
}
