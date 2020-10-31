package portadapter

// // SortServicesInfoAndApplicationServers - sort all services include child object application servers
// func SortServicesInfoAndApplicationServers(unsortedServicesInfo []*domain.ServiceInfo) []*domain.ServiceInfo {
// 	sortedServicesInfo := make([]*domain.ServiceInfo, len(unsortedServicesInfo))
// 	newServicesInfo := sortedOnlyServices(unsortedServicesInfo)
// 	for i, sortedOnlyServiceInfo := range newServicesInfo { // terrible second 'for' loop..
// 		for _, unsortedServiceInfo := range unsortedServicesInfo {
// 			if sortedOnlyServiceInfo.Address == unsortedServiceInfo.Address {
// 				sortedApplicationServers := sortApplicationServers(unsortedServiceInfo.ApplicationServers)
// 				sortedServicesInfo[i] = &domain.ServiceInfo{
// 					Address:               sortedOnlyServiceInfo.Address,
// 					IP:                    sortedOnlyServiceInfo.IP,
// 					Port:                  sortedOnlyServiceInfo.Port,
// 					IsUp:                  unsortedServiceInfo.IsUp,
// 					BalanceType:           unsortedServiceInfo.BalanceType,
// 					RoutingType:           unsortedServiceInfo.RoutingType,
// 					Protocol:              unsortedServiceInfo.Protocol,
// 					AlivedAppServersForUp: unsortedServiceInfo.AlivedAppServersForUp,
// 					HCType:                unsortedServiceInfo.HCType,
// 					HCRepeat:              unsortedServiceInfo.HCRepeat,
// 					HCTimeout:             unsortedServiceInfo.HCTimeout,
// 					HCNearFieldsMode:      unsortedServiceInfo.HCNearFieldsMode,
// 					HCUserDefinedData:     unsortedServiceInfo.HCUserDefinedData,
// 					HCRetriesForUP:        unsortedServiceInfo.HCRetriesForUP,
// 					HCRetriesForDown:      unsortedServiceInfo.HCRetriesForDown,
// 					ApplicationServers:    sortedApplicationServers,
// 					HCStop:                make(chan struct{}, 1),
// 					HCStopped:             make(chan struct{}, 1),
// 				}
// 			}
// 		}
// 	}
// 	return sortedServicesInfo
// }

// func sortedOnlyServices(unsortedServicesInfo []*domain.ServiceInfo) []*domain.ServiceInfo {
// 	servicesInfoSlice := formServicesInfoFromDomainModel(unsortedServicesInfo)
// 	onlyServicesInfoSortedSlice := sortIPs(servicesInfoSlice)
// 	return formServicesInfoDomainModelFromSlice(onlyServicesInfoSortedSlice)
// }

// func formServicesInfoFromDomainModel(servicesInfo []*domain.ServiceInfo) []string {
// 	servicesInfoSlice := make([]string, len(servicesInfo))
// 	for i, serviceInfo := range servicesInfo {
// 		servicesInfoSlice[i] = serviceInfo.Address
// 	}
// 	return servicesInfoSlice
// }

// func formServicesInfoDomainModelFromSlice(servicesInfoSlice []string) []*domain.ServiceInfo { // never check len servicesInfoSlice 2, hardcoded
// 	servicesInfo := make([]*domain.ServiceInfo, len(servicesInfoSlice))
// 	for i, serviceInfo := range servicesInfoSlice {
// 		serviceInfoSlice := strings.Split(serviceInfo, ":")
// 		servicePort := ""
// 		if len(serviceInfoSlice) > 1 {
// 			servicePort = serviceInfoSlice[1]
// 		}
// 		domainServiceInfo := &domain.ServiceInfo{
// 			Address: serviceInfoSlice[0] + ":" + servicePort,
// 			IP:      serviceInfoSlice[0],
// 			Port:    servicePort,
// 		}
// 		servicesInfo[i] = domainServiceInfo
// 	}
// 	return servicesInfo
// }

// func sortApplicationServers(applicationServers map[string]*domain.ApplicationServer) map[string]*domain.ApplicationServer {
// 	applicationServersSlice := formApplicationServersSliceFromDomainModel(applicationServers)
// 	sortedApplicationServersSlice := sortIPs(applicationServersSlice)
// 	return formApplicationServersDomainModelFromSlice(sortedApplicationServersSlice)
// }

// func formApplicationServersDomainModelFromSlice(applicationServersSlice []string) map[string]*domain.ApplicationServer { // never check len applicationServersSlice 2, hardcoded
// 	applicationServers := make(map[string]*domain.ApplicationServer, len(applicationServersSlice))
// 	for i, applicationServer := range applicationServersSlice {
// 		applicationServerSlice := strings.Split(applicationServer, ":")
// 		serverPort := ""
// 		if len(applicationServerSlice) > 1 {
// 			serverPort = applicationServerSlice[1]
// 		}
// 		domainApplicationServer := &domain.ApplicationServer{
// 			Address: applicationServerSlice[0] + ":" + serverPort,
// 			IP:      applicationServerSlice[0],
// 			Port:    serverPort,
// 		}
// 		applicationServers[i] = domainApplicationServer
// 	}
// 	return applicationServers
// }

// func formApplicationServersSliceFromDomainModel(applicationServers map[string]*domain.ApplicationServer) []string {
// 	applicationServersSlice := make([]string, len(applicationServers))
// 	for i, applicationServer := range applicationServers {
// 		applicationServersSlice[i] = applicationServer.Address
// 	}
// 	return applicationServersSlice
// }

// // sortIPs sorts IP addresses of an array in asc. order (quicksort)
// func sortIPs(addrs []string) []string {
// 	//recursion base case
// 	if len(addrs) < 2 {
// 		return addrs
// 	}
// 	pivot := addrs[rand.Intn(len(addrs))] //random ip in addrs as pivot

// 	var left []string   //for IPs<pivot
// 	var middle []string //for IPs=pivot
// 	var right []string  //for IPs>pivot

// 	for _, ip := range addrs {
// 		if orderIPPair(ip, pivot)[0] == ip {
// 			left = append(left, ip)
// 		} else if ip == pivot {
// 			middle = append(middle, ip)
// 		} else {
// 			right = append(right, ip)
// 		}
// 	}

// 	//combine and return
// 	left, right = sortIPs(left), sortIPs(right)
// 	sortedIPs := append(left, middle...)
// 	sortedIPs = append(sortedIPs, right...)
// 	return sortedIPs
// }

// // orderIPPair returns an array of two IP:PORTs in ascending order
// func orderIPPair(firstIP string, secondIP string) [2]string {
// 	// FIXME: possible panic
// 	//extract numbers out of IP strings
// 	firstIPPortArr := strings.Split(firstIP, ":")
// 	firstIPArr := strings.Split(firstIPPortArr[0], ".")
// 	secondIPPortArr := strings.Split(secondIP, ":")
// 	secondIPArr := strings.Split(secondIPPortArr[0], ".")

// 	//compare and return ordered string literal array
// 	for index, num := range firstIPArr {
// 		if num == secondIPArr[index] {
// 			continue
// 		} else if num > secondIPArr[index] {
// 			return [2]string{secondIP, firstIP}
// 		} else {
// 			return [2]string{firstIP, secondIP}
// 		}
// 	} //now check ports if nums were the same and return ordered IP:Ports
// 	if firstIPPortArr[1] < secondIPPortArr[1] {
// 		return [2]string{firstIP, secondIP}
// 	}
// 	return [2]string{secondIP, firstIP}
// }
