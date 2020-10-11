package domain

// DummyWorker ...
type DummyWorker interface {
	AddToDummy(string) error
	RemoveFromDummy(string) error
}
