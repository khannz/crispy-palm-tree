package usecase

import "github.com/khannz/crispy-palm-tree/lbost1a-dummy/domain"

type GetRuntimeConfigEntity struct {
	dummyWorker domain.DummyWorker
}

func NewGetRuntimeConfigEntity(dummyWorker domain.DummyWorker) *GetRuntimeConfigEntity {
	return &GetRuntimeConfigEntity{dummyWorker: dummyWorker}
}

func (getRuntimeConfigEntity *GetRuntimeConfigEntity) GetRuntimeConfig(id string) (map[string]struct{}, error) {
	return getRuntimeConfigEntity.dummyWorker.GetRuntimeConfig(id)
}
