package usecase

import "github.com/khannz/crispy-palm-tree/lbost1a-ipvs/domain"

type IsIPVSApplicationServerInService struct {
	ipvs domain.IPVSWorker
}

func NewIsIPVSApplicationServerInService(ipvs domain.IPVSWorker) *IsIPVSApplicationServerInService {
	return &IsIPVSApplicationServerInService{ipvs: ipvs}
}

func (isIPVSApplicationServerInService *IsIPVSApplicationServerInService) IsIPVSApplicationServerInService(serviceIP string,
	servicePort uint16,
	oneApplicationServerMap map[string]uint16,
	id string) (bool, error) {
	return isIPVSApplicationServerInService.ipvs.IsIPVSApplicationServerInService(serviceIP,
		servicePort,
		oneApplicationServerMap,
		id)
}
