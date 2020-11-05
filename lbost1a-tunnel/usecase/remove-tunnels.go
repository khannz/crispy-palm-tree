package usecase

import "github.com/khannz/crispy-palm-tree/lbost1a-tunnel/domain"

type RemoveTunnelsEntity struct {
	tunWorker domain.TunnelWorker
}

func NewRemoveTunnelsEntity(tunWorker domain.TunnelWorker) *RemoveTunnelsEntity {
	return &RemoveTunnelsEntity{tunWorker: tunWorker}
}

func (removeTunnelsEntity *RemoveTunnelsEntity) RemoveTunnels(tunnelsInfo []*domain.TunnelForApplicationServer, id string) error {
	return removeTunnelsEntity.tunWorker.RemoveTunnels(tunnelsInfo, id)
}
