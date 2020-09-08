package usecase

import (
	"fmt"

	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/sirupsen/logrus"
)

func validateRemoveApplicationServers(currentApplicattionServers,
	applicattionServersForRemove []*domain.ApplicationServer) error {
	if len(currentApplicattionServers) <= len(applicattionServersForRemove) {
		return fmt.Errorf("lenght applications servers for remove: %v. Have application servers for service: %v",
			len(applicattionServersForRemove),
			len(currentApplicattionServers))
	}

	var i []int
	for j, applicattionServerForRemove := range applicattionServersForRemove {
		for _, currentApplicattionServer := range currentApplicattionServers {
			if applicattionServerForRemove.ServerIP == currentApplicattionServer.ServerIP &&
				applicattionServerForRemove.ServerPort == currentApplicattionServer.ServerPort {
				i = append(i, j)
			}
		}
	}
	if len(i) != len(applicattionServersForRemove) {
		var formError string
		if len(i) >= 1 {
			for _, k := range i {
				formError += applicattionServersForRemove[k].ServerIP + ":" + applicattionServersForRemove[k].ServerPort + ";"
			}
			return fmt.Errorf("current config don't have application servers: %v", formError)
		}

		for _, k := range applicattionServersForRemove {
			formError += k.ServerIP + ":" + k.ServerPort + ";"
		}
		return fmt.Errorf("current config don't have application servers: %v", formError)
	}
	return nil
}

func formNewApplicationServersSlice(currentApplicattionServers, applicattionServersForRemove []*domain.ApplicationServer) []*domain.ApplicationServer {
loop:
	for i := 0; i < len(currentApplicattionServers); i++ {
		url := currentApplicattionServers[i]
		for _, rem := range applicattionServersForRemove {
			if url.ServerIP == rem.ServerIP && url.ServerPort == rem.ServerPort { // TODO: check that logic not broken from advanced healthcheck feature
				currentApplicattionServers = append(currentApplicattionServers[:i], currentApplicattionServers[i+1:]...)
				i--           // decrease index
				continue loop // TODO: labels are bad, refactor that
			}
		}
	}
	return currentApplicattionServers
}

func decreaseJobs(gracefulShutdown *domain.GracefulShutdown) {
	gracefulShutdown.Lock()
	defer gracefulShutdown.Unlock()
	gracefulShutdown.UsecasesJobs--
}

func forAddApplicationServersFormUpdateServiceInfo(currentServiceInfo, newServiceInfo *domain.ServiceInfo, eventUUID string) (*domain.ServiceInfo, error) {
	var updatedServiceInfo *domain.ServiceInfo

	// concatenate two slices
	resultApplicationServers := append(currentServiceInfo.ApplicationServers, newServiceInfo.ApplicationServers...)

	updatedServiceInfo = &domain.ServiceInfo{
		ServiceIP:          newServiceInfo.ServiceIP,
		ServicePort:        newServiceInfo.ServicePort,
		ApplicationServers: resultApplicationServers,
		Healthcheck:        currentServiceInfo.Healthcheck,
		BalanceType:        currentServiceInfo.BalanceType,
		RoutingType:        currentServiceInfo.RoutingType,
		IsUp:               currentServiceInfo.IsUp,
	}
	return updatedServiceInfo, nil
}

func forRemoveApplicationServersFormUpdateServiceInfo(currentServiceInfo, removeServiceInfo *domain.ServiceInfo, eventUUID string) *domain.ServiceInfo {
	copyOfCurrentApplicationServers := copyApplicationServers(currentServiceInfo.ApplicationServers)
	for i := 0; i < len(copyOfCurrentApplicationServers); i++ {
		if containForRemove(copyOfCurrentApplicationServers[i], removeServiceInfo.ApplicationServers) {
			copyOfCurrentApplicationServers = append(copyOfCurrentApplicationServers[:i], copyOfCurrentApplicationServers[i+1:]...)
			i--
		}
	}
	return &domain.ServiceInfo{
		ServiceIP:          currentServiceInfo.ServiceIP,
		ServicePort:        currentServiceInfo.ServicePort,
		ApplicationServers: copyOfCurrentApplicationServers,
		Healthcheck:        currentServiceInfo.Healthcheck,
		BalanceType:        currentServiceInfo.BalanceType,
		RoutingType:        currentServiceInfo.RoutingType,
		IsUp:               currentServiceInfo.IsUp,
	}
}

func copyApplicationServers(applicationServers []*domain.ApplicationServer) []*domain.ApplicationServer {
	newASs := []*domain.ApplicationServer{}
	for _, as := range applicationServers {
		copyAs := *as
		newASs = append(newASs, &copyAs)
	}
	return newASs
}

func containForRemove(tsIn *domain.ApplicationServer, toRemASs []*domain.ApplicationServer) bool {
	for _, tr := range toRemASs {
		if tsIn.ServerIP == tr.ServerIP &&
			tr.ServerPort == tr.ServerPort {
			return true
		}
	}
	return false
}

// FormTunnelsFilesInfo ...
func FormTunnelsFilesInfo(applicationServers []*domain.ApplicationServer, cacheStorage domain.StorageActions) []*domain.TunnelForApplicationServer {
	tunnelsFilesInfo := []*domain.TunnelForApplicationServer{}
	for _, applicationServer := range applicationServers {
		tunnelFilesInfo := cacheStorage.ReadTunnelInfoForApplicationServer(applicationServer.ServerIP)
		if tunnelFilesInfo == nil {
			tunnelFilesInfo = &domain.TunnelForApplicationServer{
				ApplicationServerIP:   applicationServer.ServerIP,
				ServicesToTunnelCount: 0,
			}
		}
		tunnelsFilesInfo = append(tunnelsFilesInfo, tunnelFilesInfo)
	}
	return tunnelsFilesInfo
}

func checkRoutingTypeForApplicationServersValid(newServiceInfo *domain.ServiceInfo, allCurrentServices []*domain.ServiceInfo) error {
	for _, currentService := range allCurrentServices {
		for _, newApplicationServer := range newServiceInfo.ApplicationServers {
			for _, currentApplicationServer := range currentService.ApplicationServers {
				if newApplicationServer.ServerIP == currentApplicationServer.ServerIP {
					if newServiceInfo.RoutingType != currentService.RoutingType {
						return fmt.Errorf("routing type %v for service %v for application server %v the type of routing is different from the previous routing type %v at service %v for application server %v",
							newServiceInfo.RoutingType,
							newServiceInfo.ServiceIP+":"+newServiceInfo.ServicePort,
							newApplicationServer.ServerIP+":"+newApplicationServer.ServerPort,
							currentService.RoutingType,
							currentService.ServiceIP+":"+currentService.ServicePort,
							currentApplicationServer.ServerIP+":"+currentApplicationServer.ServerPort)
					}
					continue
				}
			}
		}
	}
	return nil
}

func checkServiceIPAndPortUnique(incomeServiceIP, incomeServicePort string,
	allCurrentServices []*domain.ServiceInfo) error {
	for _, currentService := range allCurrentServices {
		if incomeServiceIP == currentService.ServiceIP && incomeServicePort == currentService.ServicePort {
			return fmt.Errorf("service %v:%v not unique: it is already in services", incomeServiceIP, incomeServicePort)
		}
		for _, currentApplicationServer := range currentService.ApplicationServers {
			if incomeServiceIP == currentApplicationServer.ServerIP && incomeServicePort == currentApplicationServer.ServerPort {
				return fmt.Errorf("service %v:%v not unique: it is already in service %v:%v as application server",
					incomeServiceIP,
					incomeServicePort,
					currentService.ServiceIP,
					currentService.ServicePort)
			}
		}
	}
	return nil
}

func checkApplicationServersIPAndPortUnique(incomeApplicationServers []*domain.ApplicationServer,
	allCurrentServices []*domain.ServiceInfo) error {
	for _, incomeApplicationServer := range incomeApplicationServers {
		for _, currentService := range allCurrentServices {
			if incomeApplicationServer.ServerIP == currentService.ServiceIP && incomeApplicationServer.ServerPort == currentService.ServicePort {
				return fmt.Errorf("application server %v:%v not unique: the same combination of ip and port is already in services",
					incomeApplicationServer.ServerIP,
					incomeApplicationServer.ServerPort)
			}
			for _, currentApplicationServer := range currentService.ApplicationServers {
				if incomeApplicationServer.ServerIP == currentApplicationServer.ServerIP && incomeApplicationServer.ServerPort == currentApplicationServer.ServerPort {
					return fmt.Errorf("application server %v:%v not unique: the same combination of ip and port at application server in service %v:%v",
						incomeApplicationServer.ServerIP,
						incomeApplicationServer.ServerPort,
						currentService.ServiceIP,
						currentService.ServicePort)
				}
			}
		}
	}
	return nil
}

func isServiceExist(incomeServiceIP, incomeServicePort string,
	allCurrentServices []*domain.ServiceInfo) bool {
	for _, currentService := range allCurrentServices {
		if incomeServiceIP == currentService.ServiceIP && incomeServicePort == currentService.ServicePort {
			return true
		}
	}
	return false
}

func checkApplicationServersExistInService(incomeApplicationServers []*domain.ApplicationServer,
	currentService *domain.ServiceInfo) error {
	for _, incomeApplicationServer := range incomeApplicationServers {
		if err := checkApplicationServerExistInService(incomeApplicationServer, currentService); err != nil {
			return err
		}
	}
	return nil
}

func checkApplicationServerExistInService(incomeApplicationServer *domain.ApplicationServer,
	currentService *domain.ServiceInfo) error {
	for _, currentApplicationServer := range currentService.ApplicationServers {
		if incomeApplicationServer.ServerIP == currentApplicationServer.ServerIP && incomeApplicationServer.ServerPort == currentApplicationServer.ServerPort {
			return nil
		}
	}
	return fmt.Errorf("application server %v:%v not finded in service %v:%v",
		incomeApplicationServer.ServerIP,
		incomeApplicationServer.ServerPort,
		currentService.ServiceIP,
		currentService.ServicePort)
}

// logging utils start TODO: move to other file log logic
func logStartUsecase(usecaseName,
	usecaseMessage,
	uuid string,
	serviceInfo *domain.ServiceInfo,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":     usecaseName,
		"event uuid": uuid,
	}).Infof("start process for %v: %v", usecaseMessage, serviceInfo)
}

func logTryCreateNewTunnels(usecaseName,
	uuid string,
	tunnelsFilesInfo []*domain.TunnelForApplicationServer,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":     usecaseName,
		"event uuid": uuid,
	}).Infof("try create new tunnels: %v", tunnelsFilesInfo)
}

func logCreatedNewTunnels(usecaseName,
	uuid string,
	tunnelsFilesInfo []*domain.TunnelForApplicationServer,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":     usecaseName,
		"event uuid": uuid,
	}).Infof("new tunnels created: %v", tunnelsFilesInfo)
}

func logTryToGetCurrentServiceInfo(usecaseName,
	uuid string,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":     usecaseName,
		"event uuid": uuid,
	}).Info("try to get current service info")
}

func logGotCurrentServiceInfo(usecaseName,
	uuid string,
	serviceInfo *domain.ServiceInfo,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":     usecaseName,
		"event uuid": uuid,
	}).Infof("got current service info: %v", serviceInfo)
}

func logTryGenerateUpdatedServiceInfo(usecaseName,
	uuid string,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":     usecaseName,
		"event uuid": uuid,
	}).Info("try to generate updated service info")
}

func logGenerateUpdatedServiceInfo(usecaseName,
	uuid string,
	serviceInfo *domain.ServiceInfo,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":     usecaseName,
		"event uuid": uuid,
	}).Infof("updated service info generated: %v", serviceInfo)
}

func logTryUpdateServiceInfoAtCache(usecaseName,
	uuid string,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":     usecaseName,
		"event uuid": uuid,
	}).Info("try update service info at cache")
}

func logUpdateServiceInfoAtCache(usecaseName,
	uuid string,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":     usecaseName,
		"event uuid": uuid,
	}).Info("service info at cache has been updated")
}

func logTryIpvsadmApplicationServers(usecaseName,
	uuid string,
	applicationServers []*domain.ApplicationServer,
	serviceIP,
	servicePort string,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":     usecaseName,
		"event uuid": uuid,
	}).Infof("try ipvsadm add application servers %v to service %v:%v",
		applicationServers,
		serviceIP,
		servicePort)
}

func logAddedIpvsadmApplicationServers(usecaseName,
	uuid string,
	applicationServers []*domain.ApplicationServer,
	serviceIP,
	servicePort string,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":     usecaseName,
		"event uuid": uuid,
	}).Infof("ipvsadm added application servers %v for service %v:%v",
		applicationServers,
		serviceIP,
		servicePort)
}

func logTryUpdateServiceInfoAtPersistentStorage(usecaseName,
	uuid string,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":     usecaseName,
		"event uuid": uuid,
	}).Info("try update service info at persistent storage")
}

func logUpdatedServiceInfoAtPersistentStorage(usecaseName,
	uuid string,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":     usecaseName,
		"event uuid": uuid,
	}).Info("service info at persistent storage has been updated")
}

func logTryGenerateCommandsForApplicationServers(usecaseName,
	uuid string,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":     usecaseName,
		"event uuid": uuid,
	}).Info("try generate commands for application servers")
}

func logGeneratedCommandsForApplicationServers(usecaseName,
	uuid string,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":     usecaseName,
		"event uuid": uuid,
	}).Info("successfully generate commands for application servers")
}

func logUpdateServiceAtHealtchecks(usecaseName,
	uuid string,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":     usecaseName,
		"event uuid": uuid,
	}).Info("update service at healtchecks")
}

func logUpdatedServiceAtHealtchecks(usecaseName,
	uuid string,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":     usecaseName,
		"event uuid": uuid,
	}).Info("service updated at healtchecks")
}

func logTryCreateIPVSService(usecaseName,
	uuid string,
	applicationServers []*domain.ApplicationServer,
	serviceIP,
	servicePort string,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":     usecaseName,
		"event uuid": uuid,
	}).Infof("try ipvsadm create service %v:%v include application servers %v",
		serviceIP,
		servicePort,
		applicationServers)
}

func logCreatedIPVSService(usecaseName,
	uuid string,
	applicationServers []*domain.ApplicationServer,
	serviceIP,
	servicePort string,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":     usecaseName,
		"event uuid": uuid,
	}).Infof("ipvsadm created service %v:%v include application servers %v",
		serviceIP,
		servicePort,
		applicationServers)
}

func logTryValidateRemoveApplicationServers(usecaseName,
	uuid string,
	applicationServers []*domain.ApplicationServer,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":     usecaseName,
		"event uuid": uuid,
	}).Infof("try validate remove application servers %v", applicationServers)
}

func logValidateRemoveApplicationServers(usecaseName,
	uuid string,
	applicationServers []*domain.ApplicationServer,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":     usecaseName,
		"event uuid": uuid,
	}).Infof("successfully validate remove application servers %v", applicationServers)
}

func logTryRemoveIpvsadmApplicationServers(usecaseName,
	uuid string,
	applicationServers []*domain.ApplicationServer,
	serviceIP,
	servicePort string,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":     usecaseName,
		"event uuid": uuid,
	}).Infof("try ipvsadm remove application servers %v for service %v:%v",
		applicationServers,
		serviceIP,
		servicePort)
}

func logRemovedIpvsadmApplicationServers(usecaseName,
	uuid string,
	applicationServers []*domain.ApplicationServer,
	serviceIP,
	servicePort string,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":     usecaseName,
		"event uuid": uuid,
	}).Infof("ipvsadm removed application servers %v for service %v:%v",
		applicationServers,
		serviceIP,
		servicePort)
}

func logTryRemoveIpvsadmService(usecaseName,
	uuid string,
	serviceInfo *domain.ServiceInfo,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":     usecaseName,
		"event uuid": uuid,
	}).Infof("try ipvsadm remove service %v", serviceInfo)
}

func logRemovedIpvsadmService(usecaseName,
	uuid string,
	serviceInfo *domain.ServiceInfo,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":     usecaseName,
		"event uuid": uuid,
	}).Infof("ipvsadm removed service %v", serviceInfo)
}

func logTryRemoveServiceAtHealtchecks(usecaseName,
	uuid string,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":     usecaseName,
		"event uuid": uuid,
	}).Info("try remove service from healtchecks")
}

func logRemovedServiceAtHealtchecks(usecaseName,
	uuid string,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":     usecaseName,
		"event uuid": uuid,
	}).Info("removed service from healtchecks")
}

func logTryRemoveIPFromDummy(usecaseName,
	uuid,
	serviceIP string,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":     usecaseName,
		"event uuid": uuid,
	}).Infof("try remove service ip %v from dummy", serviceIP)
}

func logRemovedIPFromDummy(usecaseName,
	uuid,
	serviceIP string,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":     usecaseName,
		"event uuid": uuid,
	}).Infof("removed service ip %v from dummy", serviceIP)
}

func logServicesIPAndPortNotEqual(serviceOneIP,
	serviceOnePort,
	serviceTwoIP,
	serviceTwoPort,
	usecaseName,
	uuid string,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":     usecaseName,
		"event uuid": uuid,
	}).Infof("somehow services ip and port not equal: %v:%v vs %v:%v",
		serviceOneIP,
		serviceOnePort,
		serviceTwoIP,
		serviceTwoPort)
}

func logServicesHaveDifferentNumberOfApplicationServers(serviceOneIP,
	serviceOnePort,
	serviceTwoIP,
	serviceTwoPort string,
	lenOfApplicationServersOne,
	lenOfApplicationServersTwo int,
	usecaseName,
	uuid string,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":     usecaseName,
		"event uuid": uuid,
	}).Infof("the number of application servers in the current service does not match the modification request. Services %v:%v: %v servers; %v:%v: %v servers",
		serviceOneIP,
		serviceOnePort,
		lenOfApplicationServersOne,
		serviceTwoIP,
		serviceTwoPort,
		lenOfApplicationServersTwo)
}

func logApplicationServerNotFound(serviceOneIP,
	serviceOnePort,
	applicaionServerIP,
	applicaionServerPort,
	usecaseName,
	uuid string,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":     usecaseName,
		"event uuid": uuid,
	}).Infof("in service %v:%v not found data for modify application server %v:%v",
		serviceOneIP,
		serviceOnePort,
		applicaionServerIP,
		applicaionServerPort)
}

func logTryPreValidateRequest(usecaseName,
	uuid string,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":     usecaseName,
		"event uuid": uuid,
	}).Info("start pre validation for request")
}

func logPreValidateRequestIsOk(usecaseName,
	uuid string,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":     usecaseName,
		"event uuid": uuid,
	}).Info("successfully pre validate request")
}

// logging utils end
