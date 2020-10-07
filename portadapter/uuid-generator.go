package portadapter

import "github.com/gofrs/uuid"

// UUIDGenerator ...
type UUIDGenerator struct {
}

// NewUUIIDGenerator ...
func NewUUIIDGenerator() *IDGenerator {
	return &IDGenerator{}
}

// NewID generate new ID in domain id model/struct
func (uuidGenerator *UUIDGenerator) NewID() string {
	id, _ := uuid.NewV4()
	return id.String()
}
