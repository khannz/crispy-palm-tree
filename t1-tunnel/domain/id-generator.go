package domain

// IDgenerator generates a new id
type IDgenerator interface {
	NewID() string
}
