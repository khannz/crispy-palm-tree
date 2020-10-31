package application

import "github.com/khannz/crispy-palm-tree/lbost1a-orchestrator/domain"

func convertDomainServiceInfoToRestService(domainServiceInfo *domain.ServiceInfo) *Service {
	preparedApplicationServers := convertDomainApplicationServersToRestApplicationServers(domainServiceInfo.ApplicationServers)
	restService := &Service{}
	restService.IP = domainServiceInfo.IP
	restService.Port = domainServiceInfo.Port
	restService.IsUp = domainServiceInfo.IsUp
	restService.BalanceType = domainServiceInfo.BalanceType
	restService.RoutingType = domainServiceInfo.RoutingType
	restService.Protocol = domainServiceInfo.Protocol
	restService.AlivedAppServersForUp = domainServiceInfo.AlivedAppServersForUp
	restService.HCType = domainServiceInfo.HCType
	restService.HCRepeat = domainServiceInfo.HCRepeat
	restService.HCTimeout = domainServiceInfo.HCTimeout
	restService.HCNearFieldsMode = domainServiceInfo.HCNearFieldsMode
	restService.HCUserDefinedData = domainServiceInfo.HCUserDefinedData
	restService.HCRetriesForUP = domainServiceInfo.HCRetriesForUP
	restService.HCRetriesForDown = domainServiceInfo.HCRetriesForDown
	restService.ApplicationServers = preparedApplicationServers
	return restService
}

func convertDomainApplicationServersToRestApplicationServers(domainApplicationServers map[string]*domain.ApplicationServer) []*ApplicationServer {
	preparedApplicationServers := make([]*ApplicationServer, len(domainApplicationServers))
	for _, domainApplicationServer := range domainApplicationServers {
		preparedApplicationServer := &ApplicationServer{}
		preparedApplicationServer.IP = domainApplicationServer.IP
		preparedApplicationServer.Port = domainApplicationServer.Port
		preparedApplicationServer.IsUp = domainApplicationServer.IsUp
		preparedApplicationServer.HCAddress = domainApplicationServer.HCAddress
		preparedApplicationServer.ExampleBashCommands = domainApplicationServer.ExampleBashCommands
		preparedApplicationServers = append(preparedApplicationServers, preparedApplicationServer)
	}
	return preparedApplicationServers
}

func convertRestServiceToDomainServiceInfo(restService *Service) *domain.ServiceInfo {
	preparedApplicationServers := convertRestApplicationServersToDomainApplicationServers(restService.ApplicationServers)
	domainServiceInfo := &domain.ServiceInfo{}
	domainServiceInfo.Address = restService.IP + ":" + restService.Port
	domainServiceInfo.IP = restService.IP
	domainServiceInfo.Port = restService.Port
	domainServiceInfo.IsUp = false // unknown state = false
	domainServiceInfo.BalanceType = restService.BalanceType
	domainServiceInfo.RoutingType = restService.RoutingType
	domainServiceInfo.Protocol = restService.Protocol
	domainServiceInfo.AlivedAppServersForUp = restService.AlivedAppServersForUp
	domainServiceInfo.HCType = restService.HCType
	domainServiceInfo.HCRepeat = restService.HCRepeat
	domainServiceInfo.HCTimeout = restService.HCTimeout
	domainServiceInfo.HCNearFieldsMode = restService.HCNearFieldsMode
	domainServiceInfo.HCUserDefinedData = restService.HCUserDefinedData
	domainServiceInfo.HCRetriesForUP = restService.HCRetriesForUP
	domainServiceInfo.HCRetriesForDown = restService.HCRetriesForDown
	domainServiceInfo.ApplicationServers = preparedApplicationServers
	domainServiceInfo.HCStop = make(chan struct{}, 1)
	domainServiceInfo.HCStopped = make(chan struct{}, 1)
	return domainServiceInfo
}

func convertRestApplicationServersToDomainApplicationServers(restApplicationServers []*ApplicationServer) map[string]*domain.ApplicationServer {
	preparedApplicationServers := make(map[string]*domain.ApplicationServer, len(restApplicationServers))
	for _, restApplicationServer := range restApplicationServers {
		preparedApplicationServer := &domain.ApplicationServer{}
		preparedApplicationServer.Address = restApplicationServer.IP + ":" + restApplicationServer.Port
		preparedApplicationServer.IP = restApplicationServer.IP
		preparedApplicationServer.Port = restApplicationServer.Port
		preparedApplicationServer.IsUp = false // unknown state = false
		preparedApplicationServer.HCAddress = restApplicationServer.HCAddress
		preparedApplicationServers[restApplicationServer.IP+":"+restApplicationServer.Port] = preparedApplicationServer
	}
	return preparedApplicationServers
}
