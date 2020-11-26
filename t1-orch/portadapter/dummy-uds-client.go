package portadapter

type DummyEntity struct {
}

func NewDummyEntity() *DummyEntity {
	return &DummyEntity{}
}

func (dummyEntity *DummyEntity) AddToDummy(string, string) error {
	return nil
}

func (dummyEntity *DummyEntity) RemoveFromDummy(string, string) error {
	return nil
}

func (dummyEntity *DummyEntity) GetDummyRuntimeConfig(string) (map[string]struct{}, error) {
	return nil, nil
}
