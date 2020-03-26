package domain

import "github.com/gofrs/uuid"

// UUID internal uuid structure
type UUID struct {
	UUID uuid.UUID
}

// UUIDgenerator generates a new uuid
type UUIDgenerator interface {
	NewUUID() UUID
	GetNilUUID() UUID
}
