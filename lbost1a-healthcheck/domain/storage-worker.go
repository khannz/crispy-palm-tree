package domain

// StorageActions ...
type StorageActions interface {
	NewHCServiceToStorage(*HCService, string) error
	RemoveHCServiceFromStorage(*HCService, string) error
	GetHCService(*HCService, string) (*HCService, error)
	LoadAllStorageDataToDomainModels() ([]*HCService, error)
	UpdateHCService(*HCService, string) error
}
