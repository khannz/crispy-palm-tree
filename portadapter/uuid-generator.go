package portadapter

import (
	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/gofrs/uuid"
)

// UUIDGenerator ...
type UUIDGenerator struct {
}

// NewUUIDGenerator ...
func NewUUIDGenerator() *UUIDGenerator {
	return &UUIDGenerator{}
}

// NewUUID generate new UUID in domain uuid model/struct
func (UUIDGenerator *UUIDGenerator) NewUUID() domain.UUID {
	ud, _ := uuid.NewV4()
	newUUID := domain.UUID{
		UUID: ud,
	}
	return newUUID
}

// GetNilUUID generate new UUID in domain uuid model/struct
func (UUIDGenerator *UUIDGenerator) GetNilUUID() domain.UUID {
	newUUID := domain.UUID{
		UUID: uuid.Nil,
	}
	return newUUID
}
