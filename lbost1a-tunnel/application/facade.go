package application

import (
	domain "github.com/khannz/crispy-palm-tree/lbost1a-tunnel/domain"
	"github.com/khannz/crispy-palm-tree/lbost1a-tunnel/usecase"
	"github.com/sirupsen/logrus"
)

// TunnelFacade struct
type TunnelFacade struct {
	TunnelWorker domain.TunnelWorker
	IDgenerator  domain.IDgenerator
	Logging      *logrus.Logger
}

// NewTunnelFacade ...
func NewTunnelFacade(ipvsWorker domain.TunnelWorker,
	idGenerator domain.IDgenerator,
	logging *logrus.Logger) *TunnelFacade {

	return &TunnelFacade{
		TunnelWorker: ipvsWorker,
		IDgenerator:  idGenerator,
		Logging:      logging,
	}
}

func (tunnelFacade *TunnelFacade) CreateTunnels(tunnelsInfo []*domain.TunnelForApplicationServer, id string) error {
	newCreateTunnelsEntity := usecase.NewCreateTunnelsEntity(tunnelFacade.TunnelWorker)
	return newCreateTunnelsEntity.CreateTunnels(tunnelsInfo, id)
}

func (tunnelFacade *TunnelFacade) RemoveTunnels(tunnelsInfo []*domain.TunnelForApplicationServer, id string) error {
	newRemoveTunnelsEntity := usecase.NewRemoveTunnelsEntity(tunnelFacade.TunnelWorker)
	return newRemoveTunnelsEntity.RemoveTunnels(tunnelsInfo, id)
}

func (tunnelFacade *TunnelFacade) RemoveAllTunnels(tunnelsInfo []*domain.TunnelForApplicationServer, id string) error {
	newRemoveAllTunnelsEntity := usecase.NewRemoveAllTunnelsEntity(tunnelFacade.TunnelWorker)
	return newRemoveAllTunnelsEntity.RemoveAllTunnels(tunnelsInfo, id)
}
