package application

import (
	"time"

	domain "github.com/khannz/crispy-palm-tree/lbost1a-ipvs/domain"
	"github.com/khannz/crispy-palm-tree/lbost1a-ipvs/usecase"
	"github.com/rs/zerolog"
)

type IPVSFacade struct {
	IPVSWorker         domain.IPVSWorker
	OrchestratorWorker domain.OrchestratorWorker
	Logger             *zerolog.Logger
}

func NewIPVSFacade(
	ipvsWorker domain.IPVSWorker,
	orchestratorWorker domain.OrchestratorWorker,
	logger *zerolog.Logger,
) *IPVSFacade {
	return &IPVSFacade{
		IPVSWorker:         ipvsWorker,
		OrchestratorWorker: orchestratorWorker,
		Logger:             logger,
	}
}

func (ipvsFacade *IPVSFacade) NewIPVSService(
	vip string,
	port uint16,
	routingType uint32,
	balanceType string,
	protocol uint16,
	id string,
) error {
	newNewIPVSServiceEntity := usecase.NewNewIPVSServiceEntity(ipvsFacade.IPVSWorker)
	return newNewIPVSServiceEntity.NewIPVSService(
		vip,
		port,
		routingType,
		balanceType,
		protocol,
		id,
	)
}

func (ipvsFacade *IPVSFacade) AddIPVSApplicationServersForService(
	vip string,
	port uint16,
	routingType uint32,
	balanceType string,
	protocol uint16,
	applicationServers map[string]uint16,
	id string,
) error {
	newAddApplicationServersEntity := usecase.NewAddApplicationServersEntity(ipvsFacade.IPVSWorker)
	return newAddApplicationServersEntity.AddIPVSApplicationServersForService(
		vip,
		port,
		routingType,
		balanceType,
		protocol,
		applicationServers,
		id,
	)
}

func (ipvsFacade *IPVSFacade) RemoveIPVSService(
	vip string,
	port uint16,
	protocol uint16,
	id string,
) error {
	newRemoveIPVSServiceEntity := usecase.NewRemoveIPVSServiceEntity(ipvsFacade.IPVSWorker)
	return newRemoveIPVSServiceEntity.RemoveIPVSService(
		vip,
		port,
		protocol,
		id,
	)
}

func (ipvsFacade *IPVSFacade) RemoveIPVSApplicationServersFromService(
	vip string,
	port uint16,
	routingType uint32,
	balanceType string,
	protocol uint16,
	applicationServers map[string]uint16,
	id string,
) error {
	newRemoveIPVSApplicationServersFromServiceEntity := usecase.NewRemoveIPVSApplicationServersFromServiceEntity(ipvsFacade.IPVSWorker)
	return newRemoveIPVSApplicationServersFromServiceEntity.RemoveIPVSApplicationServersFromService(
		vip,
		port,
		routingType,
		balanceType,
		protocol,
		applicationServers,
		id,
	)
}

func (ipvsFacade *IPVSFacade) GetIPVSRuntime(id string) (map[string]map[string]uint16, error) {
	newIsIPVSApplicationServerInService := usecase.NewGetIPVSRuntimeEntity(ipvsFacade.IPVSWorker)
	return newIsIPVSApplicationServerInService.GetIPVSRuntime(id)
}

// TODO Sleeps for real????
func (ipvsFacade *IPVSFacade) TryToSendRuntimeConfig(id string) {
	newGetRuntimeConfigEntity := usecase.NewGetIPVSRuntimeEntity(ipvsFacade.IPVSWorker)
	newOrchestratorSenderEntity := usecase.NewOrchestratorSenderEntity(ipvsFacade.OrchestratorWorker)
	for {
		currentConfig, err := newGetRuntimeConfigEntity.GetIPVSRuntime(id)
		if err != nil {
			ipvsFacade.Logger.Error().Err(err).Msg("failed to get runtime config")
			time.Sleep(5 * time.Second) // FIXME
			continue
		}

		if err := newOrchestratorSenderEntity.SendToOrch(currentConfig, id); err != nil {
			ipvsFacade.Logger.Error().Err(err).Msg("failed to send runtime config")
			time.Sleep(5 * time.Second) // FIXME
			continue
		}
		ipvsFacade.Logger.Info().Msg("send runtime config to hc success")
		break
	}
}
