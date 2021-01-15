package usecase

import (
	"fmt"

	"github.com/khannz/crispy-palm-tree/t1-orch/domain"
	"github.com/khannz/crispy-palm-tree/t1-orch/healthcheck"
	"github.com/sirupsen/logrus"
)

const updateServiceName = "update service"

// UpdateServiceEntity ...
type UpdateServiceEntity struct {
	memoryWorker     domain.MemoryWorker
	tunnelMaker      domain.TunnelWorker
	routeMaker       domain.RouteWorker
	ipRuleWorker     domain.IpRuleWorker
	hc               *healthcheck.HeathcheckEntity
	gracefulShutdown *domain.GracefulShutdown
	logging          *logrus.Logger
}

// NewUpdateServiceEntity ... // TODO: naming
func NewUpdateServiceEntity(memoryWorker domain.MemoryWorker,
	tunnelMaker domain.TunnelWorker,
	routeMaker domain.RouteWorker,
	ipRuleWorker domain.IpRuleWorker,
	hc *healthcheck.HeathcheckEntity,
	gracefulShutdown *domain.GracefulShutdown,
	logging *logrus.Logger) *UpdateServiceEntity {
	return &UpdateServiceEntity{
		memoryWorker:     memoryWorker,
		tunnelMaker:      tunnelMaker,
		routeMaker:       routeMaker,
		ipRuleWorker:     ipRuleWorker,
		hc:               hc,
		gracefulShutdown: gracefulShutdown,
		logging:          logging,
	}
}

// UpdateService ...
func (updateService *UpdateServiceEntity) UpdateService(serviceInfo *domain.ServiceInfo,
	updateServiceID string) error {
	// graceful shutdown part start
	updateService.gracefulShutdown.Lock()
	if updateService.gracefulShutdown.ShutdownNow {
		defer updateService.gracefulShutdown.Unlock()
		return fmt.Errorf("got shutdown signal, job new service %v canceled", serviceInfo)
	}
	updateService.gracefulShutdown.UsecasesJobs++
	updateService.gracefulShutdown.Unlock()
	defer decreaseJobs(updateService.gracefulShutdown)
	// graceful shutdown part end
	updateService.logging.WithFields(logrus.Fields{
		"entity":   updateServiceName,
		"event id": updateServiceID,
	}).Infof("start usecase for update service: %v", serviceInfo)

	// if app server exist in service do nothing
	// if not remove
	currentServiceInfo, err := updateService.memoryWorker.GetService(serviceInfo.Address)
	if err != nil {
		return err
	}
	appServersForCreate, _, appServersForRemove := updateService.formDiffForApplicationServersInfo(currentServiceInfo.ApplicationServers, serviceInfo.ApplicationServers)
	if serviceInfo.RoutingType == "tunneling" {
		for _, appSrvForAdd := range appServersForCreate {
			needCreateTunnel := updateService.memoryWorker.AddTunnelForApplicationServer(appSrvForAdd.IP)
			if needCreateTunnel {
				if err := addTunnelRouteIpRule(updateService.tunnelMaker,
					updateService.routeMaker,
					updateService.ipRuleWorker,
					serviceInfo.IP,
					appSrvForAdd.IP,
					updateServiceID); err != nil {
					return err
				}
			}
		}

		for _, appSrvForRemove := range appServersForRemove {
			needRemoveTunnel := updateService.memoryWorker.RemoveTunnelForApplicationServer(appSrvForRemove.IP)
			if needRemoveTunnel {
				if err := removeRouteTunnelIpRule(updateService.routeMaker,
					updateService.tunnelMaker,
					updateService.ipRuleWorker,
					serviceInfo.IP,
					appSrvForRemove.IP,
					updateServiceID); err != nil {
					return err
				}
			}
		}
	}

	if err := updateService.memoryWorker.UpdateService(serviceInfo); err != nil {
		return err
	}

	updateService.logging.WithFields(logrus.Fields{
		"entity":   updateServiceName,
		"event id": updateServiceID,
	}).Infof("update service %v at healtchecks", serviceInfo.Address)
	if _, err := updateService.hc.UpdateServiceAtHealtchecks(serviceInfo, updateServiceID); err != nil {
		return fmt.Errorf("error when change service in healthcheck: %v", err)
	}
	return nil
}

func (updateService *UpdateServiceEntity) formDiffForApplicationServersInfo(currentApplicationServers,
	updatedApplicationServers map[string]*domain.ApplicationServer) (map[string]*domain.ApplicationServer,
	map[string]*domain.ApplicationServer,
	map[string]*domain.ApplicationServer) {
	appServersForCreate := make(map[string]*domain.ApplicationServer)
	appServersForUpdate := make(map[string]*domain.ApplicationServer)
	appServersForRemove := make(map[string]*domain.ApplicationServer)

	for updatedAppServerInfoAddress, updatedAppServerInfo := range updatedApplicationServers {
		if _, isAppSrvIn := currentApplicationServers[updatedAppServerInfoAddress]; isAppSrvIn {
			appServersForUpdate[updatedAppServerInfoAddress] = updatedAppServerInfo
		} else {
			appServersForCreate[updatedAppServerInfoAddress] = updatedAppServerInfo
		}
	}

	for memApplicationServerAddress, memApplicationServerInfo := range currentApplicationServers {
		if _, isServiceIn := updatedApplicationServers[memApplicationServerAddress]; !isServiceIn {
			appServersForRemove[memApplicationServerAddress] = memApplicationServerInfo
		}
	}
	return appServersForCreate, appServersForUpdate, appServersForRemove
}
