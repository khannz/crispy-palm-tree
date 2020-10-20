package domain

// DummyWorker ...
type DummyWorker interface {
	AddToDummy(string, string) error
	RemoveFromDummy(string, string) error
}
