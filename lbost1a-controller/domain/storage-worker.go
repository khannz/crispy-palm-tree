package domain

// StorageActions ...
type StorageActions interface {
	NewServiceInfoToStorage(*ServiceInfo, string) error
	RemoveServiceInfoFromStorage(*ServiceInfo, string) error
	GetServiceInfo(*ServiceInfo, string) (*ServiceInfo, error)
	LoadAllStorageDataToDomainModels() ([]*ServiceInfo, error)
	UpdateServiceInfo(*ServiceInfo, string) error
	ReadTunnelInfoForApplicationServer(string) *TunnelForApplicationServer
	UpdateTunnelFilesInfoAtStorage([]*TunnelForApplicationServer) error
}
