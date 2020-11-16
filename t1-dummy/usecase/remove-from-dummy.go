package usecase

import "github.com/khannz/crispy-palm-tree/lbost1a-dummy/domain"

type RemoveFromDummyEntity struct {
	dummyWorker domain.DummyWorker
}

func NewRemoveFromDummyEntity(dummyWorker domain.DummyWorker) *RemoveFromDummyEntity {
	return &RemoveFromDummyEntity{dummyWorker: dummyWorker}
}

func (removeFromDummyEntity *RemoveFromDummyEntity) RemoveFromDummy(ip string, id string) error {
	return removeFromDummyEntity.dummyWorker.RemoveFromDummy(ip, id)
}
