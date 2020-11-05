package usecase

import "github.com/khannz/crispy-palm-tree/lbost1a-tunnel/domain"

type CreateTunnelsEntity struct {
	tunWorker domain.TunnelWorker
}

func NewCreateTunnelsEntity(tunWorker domain.TunnelWorker) *CreateTunnelsEntity {
	return &CreateTunnelsEntity{tunWorker: tunWorker}
}

func (createTunnelsEntity *CreateTunnelsEntity) CreateTunnels(tunnelsInfo []*domain.TunnelForApplicationServer, id string) error {
	return createTunnelsEntity.tunWorker.CreateTunnels(tunnelsInfo, id)
}
