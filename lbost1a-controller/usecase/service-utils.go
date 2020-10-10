package usecase

import (
	"fmt"

	"github.com/khannz/crispy-palm-tree/lbost1a-controller/domain"
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
			if applicattionServerForRemove.IP == currentApplicattionServer.IP &&
				applicattionServerForRemove.Port == currentApplicattionServer.Port {
				i = append(i, j)
			}
		}
	}
	if len(i) != len(applicattionServersForRemove) {
		var formError string
		if len(i) >= 1 {
			for _, k := range i {
				formError += applicattionServersForRemove[k].IP + ":" + applicattionServersForRemove[k].Port + ";"
			}
			return fmt.Errorf("current config don't have application servers: %v", formError)
		}

		for _, k := range applicattionServersForRemove {
			formError += k.IP + ":" + k.Port + ";"
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
			if url.IP == rem.IP && url.Port == rem.Port { // TODO: check that logic not broken from advanced healthcheck feature
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

func forAddApplicationServersFormUpdateServiceInfo(currentServiceInfo, newServiceInfo *domain.ServiceInfo, eventID string) (*domain.ServiceInfo, error) {
	var updatedServiceInfo *domain.ServiceInfo

	// concatenate two slices
	resultApplicationServers := append(currentServiceInfo.ApplicationServers, newServiceInfo.ApplicationServers...)

	updatedServiceInfo = &domain.ServiceInfo{
		Address:               newServiceInfo.Address,
		IP:                    newServiceInfo.IP,
		Port:                  newServiceInfo.Port,
		IsUp:                  currentServiceInfo.IsUp,
		BalanceType:           currentServiceInfo.BalanceType,
		RoutingType:           currentServiceInfo.RoutingType,
		Protocol:              currentServiceInfo.Protocol,
		AlivedAppServersForUp: currentServiceInfo.AlivedAppServersForUp,
		HCType:                currentServiceInfo.HCType,
		HCRepeat:              currentServiceInfo.HCRepeat,
		HCTimeout:             currentServiceInfo.HCTimeout,
		HCNearFieldsMode:      currentServiceInfo.HCNearFieldsMode,
		HCUserDefinedData:     currentServiceInfo.HCUserDefinedData,
		HCRetriesForUP:        currentServiceInfo.HCRetriesForUP,
		HCRetriesForDown:      currentServiceInfo.HCRetriesForDown,
		ApplicationServers:    resultApplicationServers,
		HCStop:                make(chan struct{}, 1),
		HCStopped:             make(chan struct{}, 1),
	}
	return updatedServiceInfo, nil
}

func forRemoveApplicationServersFormUpdateServiceInfo(currentServiceInfo, removeServiceInfo *domain.ServiceInfo, eventID string) *domain.ServiceInfo {
	copyOfCurrentApplicationServers := copyApplicationServers(currentServiceInfo.ApplicationServers)
	for i := 0; i < len(copyOfCurrentApplicationServers); i++ {
		if containForRemove(copyOfCurrentApplicationServers[i], removeServiceInfo.ApplicationServers) {
			copyOfCurrentApplicationServers = append(copyOfCurrentApplicationServers[:i], copyOfCurrentApplicationServers[i+1:]...)
			i--
		}
	}
	return &domain.ServiceInfo{
		Address:               currentServiceInfo.Address,
		IP:                    currentServiceInfo.IP,
		Port:                  currentServiceInfo.Port,
		IsUp:                  currentServiceInfo.IsUp,
		BalanceType:           currentServiceInfo.BalanceType,
		RoutingType:           currentServiceInfo.RoutingType,
		Protocol:              currentServiceInfo.Protocol,
		AlivedAppServersForUp: currentServiceInfo.AlivedAppServersForUp,
		HCType:                currentServiceInfo.HCType,
		HCRepeat:              currentServiceInfo.HCRepeat,
		HCTimeout:             currentServiceInfo.HCTimeout,
		HCNearFieldsMode:      currentServiceInfo.HCNearFieldsMode,
		HCUserDefinedData:     currentServiceInfo.HCUserDefinedData,
		HCRetriesForUP:        currentServiceInfo.HCRetriesForUP,
		HCRetriesForDown:      currentServiceInfo.HCRetriesForDown,
		ApplicationServers:    copyOfCurrentApplicationServers,
		HCStop:                make(chan struct{}, 1),
		HCStopped:             make(chan struct{}, 1),
	}
}

func copyApplicationServers(applicationServers []*domain.ApplicationServer) []*domain.ApplicationServer {
	newApplicationServers := make([]*domain.ApplicationServer, len(applicationServers))
	for i, applicationServer := range applicationServers {
		newApplicationServers[i] = &domain.ApplicationServer{
			Address:             applicationServer.Address,
			IP:                  applicationServer.IP,
			Port:                applicationServer.Port,
			IsUp:                applicationServer.IsUp,
			HCAddress:           applicationServer.HCAddress,
			ExampleBashCommands: applicationServer.ExampleBashCommands,
		}
	}
	return newApplicationServers
}

func containForRemove(tsIn *domain.ApplicationServer, toRemASs []*domain.ApplicationServer) bool {
	for _, tr := range toRemASs {
		if tsIn.IP == tr.IP &&
			tsIn.Port == tr.Port {
			return true
		}
	}
	return false
}

// FormTunnelsFilesInfo ...
func FormTunnelsFilesInfo(applicationServers []*domain.ApplicationServer, cacheStorage domain.StorageActions) []*domain.TunnelForApplicationServer {
	tunnelsFilesInfo := []*domain.TunnelForApplicationServer{}
	for _, applicationServer := range applicationServers {
		tunnelFilesInfo := cacheStorage.ReadTunnelInfoForApplicationServer(applicationServer.IP)
		if tunnelFilesInfo == nil {
			tunnelFilesInfo = &domain.TunnelForApplicationServer{
				ApplicationServerIP:   applicationServer.IP,
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
				if newApplicationServer.IP == currentApplicationServer.IP {
					if newServiceInfo.RoutingType != currentService.RoutingType {
						return fmt.Errorf("routing type %v for service %v for application server %v the type of routing is different from the previous routing type %v at service %v for application server %v",
							newServiceInfo.RoutingType,
							newServiceInfo.IP+":"+newServiceInfo.Port,
							newApplicationServer.IP+":"+newApplicationServer.Port,
							currentService.RoutingType,
							currentService.IP+":"+currentService.Port,
							currentApplicationServer.IP+":"+currentApplicationServer.Port)
					}
					continue
				}
			}
		}
	}
	return nil
}

func checkIPAndPortUnique(incomeServiceIP, incomeServicePort string,
	allCurrentServices []*domain.ServiceInfo) error {
	for _, currentService := range allCurrentServices {
		if incomeServiceIP == currentService.IP && incomeServicePort == currentService.Port {
			return fmt.Errorf("service %v:%v not unique: it is already in services", incomeServiceIP, incomeServicePort)
		}
		for _, currentApplicationServer := range currentService.ApplicationServers {
			if incomeServiceIP == currentApplicationServer.IP && incomeServicePort == currentApplicationServer.Port {
				return fmt.Errorf("service %v:%v not unique: it is already in service %v:%v as application server",
					incomeServiceIP,
					incomeServicePort,
					currentService.IP,
					currentService.Port)
			}
		}
	}
	return nil
}

func checkApplicationServersIPAndPortUnique(incomeApplicationServers []*domain.ApplicationServer,
	allCurrentServices []*domain.ServiceInfo) error {
	for _, incomeApplicationServer := range incomeApplicationServers {
		for _, currentService := range allCurrentServices {
			if incomeApplicationServer.IP == currentService.IP && incomeApplicationServer.Port == currentService.Port {
				return fmt.Errorf("application server %v:%v not unique: the same combination of ip and port is already in services",
					incomeApplicationServer.IP,
					incomeApplicationServer.Port)
			}
			for _, currentApplicationServer := range currentService.ApplicationServers {
				if incomeApplicationServer.IP == currentApplicationServer.IP && incomeApplicationServer.Port == currentApplicationServer.Port {
					return fmt.Errorf("application server %v:%v not unique: the same combination of ip and port at application server in service %v:%v",
						incomeApplicationServer.IP,
						incomeApplicationServer.Port,
						currentService.IP,
						currentService.Port)
				}
			}
		}
	}
	return nil
}

func isServiceExist(incomeServiceIP, incomeServicePort string,
	allCurrentServices []*domain.ServiceInfo) bool {
	for _, currentService := range allCurrentServices {
		if incomeServiceIP == currentService.IP && incomeServicePort == currentService.Port {
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
		if incomeApplicationServer.IP == currentApplicationServer.IP && incomeApplicationServer.Port == currentApplicationServer.Port {
			return nil
		}
	}
	return fmt.Errorf("application server %v:%v not finded in service %v:%v",
		incomeApplicationServer.IP,
		incomeApplicationServer.Port,
		currentService.IP,
		currentService.Port)
}

// logging utils start TODO: move to other file log logic
func logStartUsecase(usecaseName,
	usecaseMessage,
	id string,
	serviceInfo *domain.ServiceInfo,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":   usecaseName,
		"event id": id,
	}).Infof("start process for %v: %v", usecaseMessage, serviceInfo)
}

func logTryCreateNewTunnels(usecaseName,
	id string,
	tunnelsFilesInfo []*domain.TunnelForApplicationServer,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":   usecaseName,
		"event id": id,
	}).Infof("try create new tunnels: %v", tunnelsFilesInfo)
}

func logTryRemoveTunnels(usecaseName,
	id string,
	tunnelsFilesInfo []*domain.TunnelForApplicationServer,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":   usecaseName,
		"event id": id,
	}).Infof("try remove tunnels: %v", tunnelsFilesInfo)
}

func logCreatedNewTunnels(usecaseName,
	id string,
	tunnelsFilesInfo []*domain.TunnelForApplicationServer,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":   usecaseName,
		"event id": id,
	}).Infof("new tunnels created: %v", tunnelsFilesInfo)
}

func logRemovedTunnels(usecaseName,
	id string,
	tunnelsFilesInfo []*domain.TunnelForApplicationServer,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":   usecaseName,
		"event id": id,
	}).Infof("tunnels removed: %v", tunnelsFilesInfo)
}

func logTryToGetCurrentServiceInfo(usecaseName,
	id string,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":   usecaseName,
		"event id": id,
	}).Info("try to get current service info")
}

func logGotCurrentServiceInfo(usecaseName,
	id string,
	serviceInfo *domain.ServiceInfo,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":   usecaseName,
		"event id": id,
	}).Infof("got current service info: %v", serviceInfo)
}

func logTryGenerateUpdatedServiceInfo(usecaseName,
	id string,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":   usecaseName,
		"event id": id,
	}).Info("try to generate updated service info")
}

func logGenerateUpdatedServiceInfo(usecaseName,
	id string,
	serviceInfo *domain.ServiceInfo,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":   usecaseName,
		"event id": id,
	}).Infof("updated service info generated: %v", serviceInfo)
}

func logTryUpdateServiceInfoAtCache(usecaseName,
	id string,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":   usecaseName,
		"event id": id,
	}).Info("try update service info at cache")
}

func logUpdateServiceInfoAtCache(usecaseName,
	id string,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":   usecaseName,
		"event id": id,
	}).Info("service info at cache has been updated")
}

func logTryUpdateServiceInfoAtPersistentStorage(usecaseName,
	id string,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":   usecaseName,
		"event id": id,
	}).Info("try update service info at persistent storage")
}

func logUpdatedServiceInfoAtPersistentStorage(usecaseName,
	id string,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":   usecaseName,
		"event id": id,
	}).Info("service info at persistent storage has been updated")
}

func logTryGenerateCommandsForApplicationServers(usecaseName,
	id string,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":   usecaseName,
		"event id": id,
	}).Info("try generate commands for application servers")
}

func logGeneratedCommandsForApplicationServers(usecaseName,
	id string,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":   usecaseName,
		"event id": id,
	}).Info("successfully generate commands for application servers")
}

func logUpdateServiceAtHealtchecks(usecaseName,
	id string,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":   usecaseName,
		"event id": id,
	}).Info("update service at healtchecks")
}

func logUpdatedServiceAtHealtchecks(usecaseName,
	id string,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":   usecaseName,
		"event id": id,
	}).Info("service updated at healtchecks")
}

func logTryValidateRemoveApplicationServers(usecaseName,
	id string,
	applicationServers []*domain.ApplicationServer,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":   usecaseName,
		"event id": id,
	}).Infof("try validate remove application servers %v", applicationServers)
}

func logValidateRemoveApplicationServers(usecaseName,
	id string,
	applicationServers []*domain.ApplicationServer,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":   usecaseName,
		"event id": id,
	}).Infof("successfully validate remove application servers %v", applicationServers)
}

func logTryRemoveServiceAtHealtchecks(usecaseName,
	id string,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":   usecaseName,
		"event id": id,
	}).Info("try remove service from healtchecks")
}

func logRemovedServiceAtHealtchecks(usecaseName,
	id string,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":   usecaseName,
		"event id": id,
	}).Info("removed service from healtchecks")
}

func logServicesIPAndPortNotEqual(serviceOneIP,
	serviceOnePort,
	serviceTwoIP,
	serviceTwoPort,
	usecaseName,
	id string,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":   usecaseName,
		"event id": id,
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
	id string,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":   usecaseName,
		"event id": id,
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
	id string,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":   usecaseName,
		"event id": id,
	}).Infof("in service %v:%v not found data for modify application server %v:%v",
		serviceOneIP,
		serviceOnePort,
		applicaionServerIP,
		applicaionServerPort)
}

func logTryPreValidateRequest(usecaseName,
	id string,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":   usecaseName,
		"event id": id,
	}).Info("start pre validation for request")
}

func logPreValidateRequestIsOk(usecaseName,
	id string,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":   usecaseName,
		"event id": id,
	}).Info("successfully pre validate request")
}

// logging utils end