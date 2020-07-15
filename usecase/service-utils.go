package usecase

import (
	"fmt"

	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/khannz/crispy-palm-tree/portadapter"
	"github.com/sirupsen/logrus"
)

// TODO: need better check unique, app srv to services too
func checkNewApplicationServersIsUnique(currentServiceInfo, newServiceInfo *domain.ServiceInfo, eventUUID string) error {
	// TODO: bad loops
	for _, newApplicationServer := range newServiceInfo.ApplicationServers {
		for _, currentApplicationServer := range currentServiceInfo.ApplicationServers {
			if newApplicationServer == currentApplicationServer {
				return fmt.Errorf("application server %v:%v alredy exist in service %v:%v",
					newApplicationServer.ServerIP,
					newApplicationServer.ServerPort,
					newServiceInfo.ServiceIP,
					newServiceInfo.ServicePort)
			}
		}
	}
	return nil
}

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
		for _, k := range i {
			formError += applicattionServersForRemove[k].ServerIP + ":" + applicattionServersForRemove[k].ServerPort + ";"
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
			if *url == *rem {
				currentApplicattionServers = append(currentApplicattionServers[:i], currentApplicattionServers[i+1:]...)
				i-- // decrease index
				continue loop
			}
		}
	}
	return currentApplicattionServers
}

func decreaseJobs(gracefullShutdown *domain.GracefullShutdown) {
	gracefullShutdown.Lock()
	defer gracefullShutdown.Unlock()
	gracefullShutdown.UsecasesJobs--
}

// // need to be sure fullApplicationServersInfo contain incompleteApplicationServersInfo
// func enrichApplicationServersInfo(fullApplicationServersInfo []*domain.ApplicationServer,
// 	incompleteApplicationServersInfo []*domain.ApplicationServer) []*domain.ApplicationServer {
// 	enrichApplicationServersInfo := []*domain.ApplicationServer{}
// 	for _, incompleteApplicationServerInfo := range incompleteApplicationServersInfo {
// 		for _, fullApplicationServerInfo := range fullApplicationServersInfo {
// 			if incompleteApplicationServerInfo.ServerIP == fullApplicationServerInfo.ServerIP &&
// 				incompleteApplicationServerInfo.ServerPort == fullApplicationServerInfo.ServerPort {
// 				enrichApplicationServerInfo := &domain.ApplicationServer{
// 					ServerIP:          incompleteApplicationServerInfo.ServerIP,
// 					ServerPort:        incompleteApplicationServerInfo.ServerPort,
// 					IsUp:             fullApplicationServerInfo.IsUp,
// 					IfcfgTunnelFile:   fullApplicationServerInfo.IfcfgTunnelFile,
// 					RouteTunnelFile:   fullApplicationServerInfo.RouteTunnelFile,
// 					SysctlConfFile:    fullApplicationServerInfo.SysctlConfFile,
// 					TunnelName:        fullApplicationServerInfo.TunnelName,
// 					ServerHealthcheck: fullApplicationServerInfo.ServerHealthcheck,
// 				}
// 				enrichApplicationServersInfo = append(enrichApplicationServersInfo, enrichApplicationServerInfo)
// 			}
// 		}
// 	}
// 	return enrichApplicationServersInfo
// }

func forAddApplicationServersFormUpdateServiceInfo(currentServiceInfo, newServiceInfo *domain.ServiceInfo, eventUUID string) (*domain.ServiceInfo, error) {
	var resultServiceInfo *domain.ServiceInfo
	if err := checkNewApplicationServersIsUnique(currentServiceInfo, newServiceInfo, eventUUID); err != nil {
		return resultServiceInfo, fmt.Errorf("new application server not unique: %v", err)
	}
	// concatenate two slices
	resultApplicationServers := append(currentServiceInfo.ApplicationServers, newServiceInfo.ApplicationServers...)

	resultServiceInfo = &domain.ServiceInfo{
		ServiceIP:          newServiceInfo.ServiceIP,
		ServicePort:        newServiceInfo.ServicePort,
		ApplicationServers: resultApplicationServers,
		Healthcheck:        currentServiceInfo.Healthcheck,
		IsUp:               currentServiceInfo.IsUp,
	}
	return resultServiceInfo, nil
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

// func findApplicationServeInCurrentApplicationServers(serverIP, serverPort string, currentServiceInfo *domain.ServiceInfo) (int, bool) {
// 	var findedIndex int
// 	var isFinded bool
// 	for index, currentApplicationServer := range currentServiceInfo.ApplicationServers {
// 		if serverIP == currentApplicationServer.ServerIP &&
// 			serverPort == currentApplicationServer.ServerPort {
// 			findedIndex = index
// 			isFinded = true
// 			break
// 		}
// 	}
// 	return findedIndex, isFinded
// }

func formTunnelsFilesInfo(applicationServers []*domain.ApplicationServer, cacheStorage *portadapter.StorageEntity) []*domain.TunnelForApplicationServer {
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

// logging utils start
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
	}).Info("ipvsadm added application servers %v for service %v:%v",
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
	}).Info("try ipvsadm remove application servers %v for service %v:%v",
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
	}).Info("ipvsadm removed application servers %v for service %v:%v",
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
	}).Info("try ipvsadm remove service %v", serviceInfo)
}

func logRemovedIpvsadmService(usecaseName,
	uuid string,
	serviceInfo *domain.ServiceInfo,
	logging *logrus.Logger) {
	logging.WithFields(logrus.Fields{
		"entity":     usecaseName,
		"event uuid": uuid,
	}).Info("ipvsadm removed service %v", serviceInfo)
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

// logging utils end
