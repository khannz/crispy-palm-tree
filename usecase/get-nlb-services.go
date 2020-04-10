package usecase

// import (
// 	"fmt"

// 	"github.com/khannz/crispy-palm-tree/domain"
// 	"github.com/sirupsen/logrus"
// )

// const getNlbServicesEntity = "get-nlb-services"

// // GetNlbServices ...
// type GetNlbServices struct {
// 	nwConfig         *domain.NetworkConfig
// 	keepalivedConfig domain.KeepalivedCustomizer
// 	logging          *logrus.Logger
// }

// // NewGetNlbServices ...
// func NewGetNlbServices(nwConfig *domain.NetworkConfig,
// 	keepalivedConfig domain.KeepalivedCustomizer,
// 	logging *logrus.Logger) *GetNlbServices {
// 	return &GetNlbServices{
// 		nwConfig:         nwConfig,
// 		keepalivedConfig: keepalivedConfig,
// 		logging:          logging,
// 	}
// }

// // GetAllNWBServices ...
// func (getNlbServices *GetNlbServices) GetAllNWBServices(newNWBRequestUUID string) ([]domain.ServiceInfo, error) {
// 	getNlbServices.nwConfig.Lock()
// 	defer getNlbServices.nwConfig.Unlock()
// 	servicesInfo, err := getNlbServices.keepalivedConfig.GetInfoAboutAllNWBServices(newNWBRequestUUID)
// 	if err != nil {
// 		return nil, fmt.Errorf("can't g et all nwb services: %v", err)
// 	}
// 	return servicesInfo, nil
// }
