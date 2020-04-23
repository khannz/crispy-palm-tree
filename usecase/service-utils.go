package usecase

import (
	"fmt"

	"github.com/khannz/crispy-palm-tree/domain"
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
	applicattionServersForRemove []domain.ApplicationServer) error {
	if len(currentApplicattionServers) <= len(applicattionServersForRemove) {
		return fmt.Errorf("lenght applications servers for remove: %v. Have application servers for service: %v",
			len(applicattionServersForRemove),
			len(currentApplicattionServers))
	}

	var i []int
	for j, applicattionServerForRemove := range applicattionServersForRemove {
		for _, currentApplicattionServer := range currentApplicattionServers {
			if applicattionServerForRemove == currentApplicattionServer {
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

func formNewApplicationServersSlice(currentApplicattionServers, applicattionServersForRemove []domain.ApplicationServer) []domain.ApplicationServer {
loop:
	for i := 0; i < len(currentApplicattionServers); i++ {
		url := currentApplicattionServers[i]
		for _, rem := range applicattionServersForRemove {
			if url == rem {
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

// need to be sure fullApplicationServersInfo contain incompleteApplicationServersInfo
func enrichApplicationServersInfo(fullApplicationServersInfo []domain.ApplicationServer,
	incompleteApplicationServersInfo []domain.ApplicationServer) []domain.ApplicationServer {
	enrichApplicationServersInfo := []domain.ApplicationServer{}
	for _, incompleteApplicationServerInfo := range incompleteApplicationServersInfo {
		for _, fullApplicationServerInfo := range fullApplicationServersInfo {
			if incompleteApplicationServerInfo.ServerIP == fullApplicationServerInfo.ServerIP &&
				incompleteApplicationServerInfo.ServerPort == fullApplicationServerInfo.ServerPort {
				enrichApplicationServerInfo := domain.ApplicationServer{
					ServerIP:        incompleteApplicationServerInfo.ServerIP,
					ServerPort:      incompleteApplicationServerInfo.ServerPort,
					State:           fullApplicationServerInfo.State,
					IfcfgTunnelFile: fullApplicationServerInfo.IfcfgTunnelFile,
					RouteTunnelFile: fullApplicationServerInfo.RouteTunnelFile,
					SysctlConfFile:  fullApplicationServerInfo.SysctlConfFile,
					TunnelName:      fullApplicationServerInfo.TunnelName,
				}
				enrichApplicationServersInfo = append(enrichApplicationServersInfo, enrichApplicationServerInfo)
			}
		}
	}
	return enrichApplicationServersInfo
}
