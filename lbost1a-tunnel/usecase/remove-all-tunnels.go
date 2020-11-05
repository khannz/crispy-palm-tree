package usecase

import "github.com/khannz/crispy-palm-tree/lbost1a-tunnel/domain"

type RemoveAllTunnelsEntity struct {
	tunWorker domain.TunnelWorker
}

func NewRemoveAllTunnelsEntity(tunWorker domain.TunnelWorker) *RemoveAllTunnelsEntity {
	return &RemoveAllTunnelsEntity{tunWorker: tunWorker}
}

func (removeAllTunnelsEntity *RemoveAllTunnelsEntity) RemoveAllTunnels(tunnelsInfo []*domain.TunnelForApplicationServer, id string) error {
	return removeAllTunnelsEntity.tunWorker.RemoveAllTunnels(tunnelsInfo, id)
}
