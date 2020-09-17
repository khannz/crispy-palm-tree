package portadapter

import (
	"math/rand"
	"strings"

	"github.com/khannz/crispy-palm-tree/domain"
)

// SortServicesInfoAndApplicationServers - sort all services include child object application servers
func SortServicesInfoAndApplicationServers(unsortedServicesInfo []*domain.ServiceInfo) []*domain.ServiceInfo {
	sortedServicesInfo := []*domain.ServiceInfo{}
	newServicesInfo := sortedOnlyServices(unsortedServicesInfo)
	for _, sortedOnlyServiceInfo := range newServicesInfo { // terrible second 'for' loop..
		for _, unsortedServiceInfo := range unsortedServicesInfo {
			if sortedOnlyServiceInfo.ServiceIP == unsortedServiceInfo.ServiceIP &&
				sortedOnlyServiceInfo.ServicePort == unsortedServiceInfo.ServicePort {
				sortedApplicationServers := sortApplicationServers(unsortedServiceInfo.ApplicationServers)
				sortedServiceInfo := &domain.ServiceInfo{
					ServiceIP:          sortedOnlyServiceInfo.ServiceIP,
					ServicePort:        sortedOnlyServiceInfo.ServicePort,
					ApplicationServers: sortedApplicationServers,
					Healthcheck:        unsortedServiceInfo.Healthcheck,
					BalanceType:        unsortedServiceInfo.BalanceType,
					RoutingType:        unsortedServiceInfo.RoutingType,
					ExtraInfo:          unsortedServiceInfo.ExtraInfo,
					Protocol:           unsortedServiceInfo.Protocol,
				}
				sortedServicesInfo = append(sortedServicesInfo, sortedServiceInfo)
			}
		}

	}
	return sortedServicesInfo
}

func sortedOnlyServices(unsortedServicesInfo []*domain.ServiceInfo) []*domain.ServiceInfo {
	servicesInfoSlice := formServicesInfoFromDomainModel(unsortedServicesInfo)
	onlyServicesInfoSortedSlice := sortIPs(servicesInfoSlice)
	return formServicesInfoDomainModelFromSlice(onlyServicesInfoSortedSlice)
}

func formServicesInfoFromDomainModel(servicesInfo []*domain.ServiceInfo) []string {
	servicesInfoSlice := []string{}
	for _, serviceInfo := range servicesInfo {
		servicesInfoSlice = append(servicesInfoSlice, serviceInfo.ServiceIP+":"+serviceInfo.ServicePort)
	}
	return servicesInfoSlice
}

func formServicesInfoDomainModelFromSlice(servicesInfoSlice []string) []*domain.ServiceInfo { // never check len servicesInfoSlice 2, hardcoded
	servicesInfo := []*domain.ServiceInfo{}
	for _, serviceInfo := range servicesInfoSlice {
		serviceInfoSlice := strings.Split(serviceInfo, ":")
		servicePort := ""
		if len(serviceInfoSlice) > 1 {
			servicePort = serviceInfoSlice[1]
		}
		serviceInfo := &domain.ServiceInfo{
			ServiceIP:   serviceInfoSlice[0],
			ServicePort: servicePort,
		}
		servicesInfo = append(servicesInfo, serviceInfo)
	}
	return servicesInfo
}

func sortApplicationServers(applicationServers []*domain.ApplicationServer) []*domain.ApplicationServer {
	applicationServersSlice := formApplicationServersSliceFromDomainModel(applicationServers)
	sortedApplicationServersSlice := sortIPs(applicationServersSlice)
	return formApplicationServersDomainModelFromSlice(sortedApplicationServersSlice)
}

func formApplicationServersDomainModelFromSlice(applicationServersSlice []string) []*domain.ApplicationServer { // never check len applicationServersSlice 2, hardcoded
	applicationServers := []*domain.ApplicationServer{}
	for _, applicationServer := range applicationServersSlice {
		applicationServerSlice := strings.Split(applicationServer, ":")
		serverPort := ""
		if len(applicationServerSlice) > 1 {
			serverPort = applicationServerSlice[1]
		}
		applicationServer := &domain.ApplicationServer{
			ServerIP:   applicationServerSlice[0],
			ServerPort: serverPort,
		}
		applicationServers = append(applicationServers, applicationServer)
	}
	return applicationServers
}

func formApplicationServersSliceFromDomainModel(applicationServers []*domain.ApplicationServer) []string {
	applicationServersSlice := []string{}
	for _, applicationServer := range applicationServers {
		applicationServersSlice = append(applicationServersSlice, applicationServer.ServerIP+":"+applicationServer.ServerPort)
	}
	return applicationServersSlice
}

// sortIPs sorts IP addresses of an array in asc. order (quicksort)
func sortIPs(addrs []string) []string {
	//recursion base case
	if len(addrs) < 2 {
		return addrs
	}
	pivot := addrs[rand.Intn(len(addrs))] //random ip in addrs as pivot

	var left []string   //for IPs<pivot
	var middle []string //for IPs=pivot
	var right []string  //for IPs>pivot

	for _, ip := range addrs {
		if orderIPPair(ip, pivot)[0] == ip {
			left = append(left, ip)
		} else if ip == pivot {
			middle = append(middle, ip)
		} else {
			right = append(right, ip)
		}
	}

	//combine and return
	left, right = sortIPs(left), sortIPs(right)
	sortedIPs := append(left, middle...)
	sortedIPs = append(sortedIPs, right...)
	return sortedIPs
}

// orderIPPair returns an array of two IP:PORTs in ascending order
func orderIPPair(firstIP string, secondIP string) [2]string {
	// FIXME: possible panic
	//extract numbers out of IP strings
	firstIPPortArr := strings.Split(firstIP, ":")
	firstIPArr := strings.Split(firstIPPortArr[0], ".")
	secondIPPortArr := strings.Split(secondIP, ":")
	secondIPArr := strings.Split(secondIPPortArr[0], ".")

	//compare and return ordered string literal array
	for index, num := range firstIPArr {
		if num == secondIPArr[index] {
			continue
		} else if num > secondIPArr[index] {
			return [2]string{secondIP, firstIP}
		} else {
			return [2]string{firstIP, secondIP}
		}
	} //now check ports if nums were the same and return ordered IP:Ports
	if firstIPPortArr[1] < secondIPPortArr[1] {
		return [2]string{firstIP, secondIP}
	}
	return [2]string{secondIP, firstIP}
}
