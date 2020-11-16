package portadapter

import (
	gonanoid "github.com/matoous/go-nanoid"
)

// IDGenerator ...
type IDGenerator struct {
}

// NewIDGenerator ...
func NewIDGenerator() *IDGenerator {
	return &IDGenerator{}
}

// NewID generate new ID in domain id model/struct
func (idGenerator *IDGenerator) NewID() string {
	id, _ := gonanoid.Nanoid()
	return id
}
