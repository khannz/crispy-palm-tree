package usecase

import "github.com/khannz/crispy-palm-tree/lbost1a-dummy/domain"

type AddToDummyEntity struct {
	dummyWorker domain.DummyWorker
}

func NewAddToDummyEntity(dummyWorker domain.DummyWorker) *AddToDummyEntity {
	return &AddToDummyEntity{dummyWorker: dummyWorker}
}

func (addApplicationServersEntity *AddToDummyEntity) AddToDummy(ip string) error {
	return nil
}
