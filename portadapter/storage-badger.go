package portadapter

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"
	"time"

	badger "github.com/dgraph-io/badger/v2"
	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/sirupsen/logrus"
)

// StorageEntity ...
type StorageEntity struct {
	sync.Mutex
	Db      *badger.DB
	logging *logrus.Logger
}

// NewStorageEntity ...
// do not forget defer Db.Close()!
func NewStorageEntity(isInMemory bool, dbPath string, logging *logrus.Logger) (*StorageEntity, error) {
	fakeLogging := logrus.New()
	fakeLogging.SetOutput(ioutil.Discard)
	var opt badger.Options
	if isInMemory {
		opt = optionsForDbInMemory(fakeLogging)
	} else {
		opt = optionsForDbPersistent(dbPath, fakeLogging)
	}
	db, err := badger.Open(opt)
	if err != nil {
		return nil, fmt.Errorf("can't open storage: %v", err)
	}
	storageEntity := &StorageEntity{
		Db:      db,
		logging: logging,
	}
	return storageEntity, nil
}

func optionsForDbInMemory(logger *logrus.Logger) badger.Options {
	defOpt := badger.DefaultOptions("")
	defOpt.Logger = logger
	defOpt.InMemory = true
	return defOpt
}

func optionsForDbPersistent(dbPath string, logger *logrus.Logger) badger.Options {
	defOpt := badger.DefaultOptions(dbPath)
	defOpt.Logger = logger
	return defOpt
}

// ExtendedServiceData have application servers info and info abou service
type ExtendedServiceData struct {
	ServiceRepeatHealthcheck                          time.Duration              `json:"serviceRepeatHealthcheck"`
	ServicePercentOfAlivedForUp                       int                        `json:"servicePercentOfAlivedForUp"`
	ServiceHealthcheckType                            string                     `json:"serviceHealthcheckType"`
	ServiceHealthcheckTimeout                         time.Duration              `json:"serviceHealthcheckTimeout"`
	ServiceHealthcheckRetriesForUpApplicationServer   int                        `json:"serviceHealthcheckRetriesForUpApplicationServer"`
	ServiceHealthcheckRetriesForDownApplicationServer int                        `json:"serviceHealthcheckRetriesForDownApplicationServer"`
	ServiceExtraInfo                                  []string                   `json:"serviceExtraInfo"`
	ServiceIsUp                                       bool                       `json:"serviceIsUp"`
	ApplicationServers                                []domain.ApplicationServer `json:"applicationServers"`
	BalanceType                                       string                     `json:"balanceType"`
	RoutingType                                       string                     `json:"routingType"`
	Protocol                                          string                     `json:"protocol"`
}

// TunnelForService ...
type TunnelForService struct {
	SysctlConfFile        string `json:"sysctlConf"` // full path to sysctl conf file
	TunnelName            string `json:"tunnelName"`
	ServicesToTunnelCount int    `json:"servicesToTunnelCount"`
}

// NewServiceInfoToStorage add new service to storage
func (storageEntity *StorageEntity) NewServiceInfoToStorage(serviceData *domain.ServiceInfo,
	eventUUID string) error {
	serviceDataKey, serviceDataValue, err := transformServiceDataForStorageData(serviceData)
	if err != nil {
		return fmt.Errorf("can't form data for storage: %v", err)
	}

	storageEntity.Lock()
	defer storageEntity.Unlock()

	err = storageEntity.updateDatabaseServiceInfo(serviceDataKey, serviceDataValue)
	if err != nil {
		return fmt.Errorf("can't update storage for new service: %v", err)
	}

	return nil
}

func transformServiceDataForStorageData(serviceData *domain.ServiceInfo) ([]byte,
	[]byte,
	error) {
	serviceDataKey := []byte(serviceData.ServiceIP + ":" + serviceData.ServicePort)

	renewApplicationServers := []domain.ApplicationServer{}
	for _, applicationServer := range serviceData.ApplicationServers {
		renewApplicationServer := *applicationServer
		renewApplicationServers = append(renewApplicationServers, renewApplicationServer)
	}

	transformedServiceData := ExtendedServiceData{
		ServiceRepeatHealthcheck:                          serviceData.Healthcheck.RepeatHealthcheck,
		ServicePercentOfAlivedForUp:                       serviceData.Healthcheck.PercentOfAlivedForUp,
		ServiceHealthcheckType:                            serviceData.Healthcheck.Type,
		ServiceHealthcheckTimeout:                         serviceData.Healthcheck.Timeout,
		ServiceHealthcheckRetriesForUpApplicationServer:   serviceData.Healthcheck.RetriesForUpApplicationServer,
		ServiceHealthcheckRetriesForDownApplicationServer: serviceData.Healthcheck.RetriesForDownApplicationServer,
		ServiceExtraInfo:                                  serviceData.ExtraInfo,
		ServiceIsUp:                                       serviceData.IsUp,
		ApplicationServers:                                renewApplicationServers,
		BalanceType:                                       serviceData.BalanceType,
		RoutingType:                                       serviceData.RoutingType,
		Protocol:                                          serviceData.Protocol,
	}
	serviceDataValue, err := json.Marshal(transformedServiceData)
	if err != nil {
		return nil, nil, fmt.Errorf("can't marshal transformedServiceData: %v", err)
	}
	return serviceDataKey, serviceDataValue, nil
}

func updateDb(db *badger.DB, key, value []byte) error {
	return db.Update(func(txn *badger.Txn) error {
		err := txn.Set(key, value)
		return err
	})
}

func (storageEntity *StorageEntity) updateDatabaseServiceInfo(serviceDataKey,
	serviceDataValue []byte) error {
	err := updateDb(storageEntity.Db, serviceDataKey, serviceDataValue)
	if err != nil {
		return fmt.Errorf("can't update db for service data: %v", err)
	}
	return nil
}

// RemoveServiceInfoFromStorage ...
func (storageEntity *StorageEntity) RemoveServiceInfoFromStorage(serviceData *domain.ServiceInfo, eventUUID string) error {
	keyData := []byte(serviceData.ServiceIP + ":" + serviceData.ServicePort)
	storageEntity.Lock()
	defer storageEntity.Unlock()

	if err := storageEntity.Db.Update(func(txn *badger.Txn) error {
		if err := txn.Delete(keyData); err != nil {
			return fmt.Errorf("txn.Delete fail: %v", err)
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

// GetServiceInfo ...
func (storageEntity *StorageEntity) GetServiceInfo(incomeServiceData *domain.ServiceInfo, eventUUID string) (*domain.ServiceInfo, error) {
	shc := domain.ServiceHealthcheck{
		RepeatHealthcheck:               3000000009,
		Type:                            "",
		Timeout:                         time.Duration(999 * time.Second),
		RetriesForUpApplicationServer:   0,
		RetriesForDownApplicationServer: 0,
	}
	currentServiceInfo := &domain.ServiceInfo{
		ServiceIP:          "",
		ServicePort:        "",
		ApplicationServers: []*domain.ApplicationServer{},
		Healthcheck:        shc,
		ExtraInfo:          []string{},
		IsUp:               false,
		BalanceType:        "",
		RoutingType:        "",
		Protocol:           "",
	}
	currentApplicationServers := []*domain.ApplicationServer{}
	storageEntity.Lock()
	defer storageEntity.Unlock()
	if err := storageEntity.Db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(incomeServiceData.ServiceIP + ":" + incomeServiceData.ServicePort))
		if err != nil {
			return fmt.Errorf("txn.Get fail: %v", err)
		}

		oldExtendedServiceData := ExtendedServiceData{}
		if err = item.Value(func(val []byte) error {
			if err := json.Unmarshal(val, &oldExtendedServiceData); err != nil {
				return fmt.Errorf("can't unmarshall application servers data: %v", err)
			}
			return nil
		}); err != nil {
			return err
		}
		for _, oldApplicationServer := range oldExtendedServiceData.ApplicationServers {
			renewPointerForOldApplicationServer := oldApplicationServer
			currentApplicationServers = append(currentApplicationServers, &renewPointerForOldApplicationServer)
		}

		currentServiceInfo.Healthcheck = domain.ServiceHealthcheck{
			RepeatHealthcheck:               oldExtendedServiceData.ServiceRepeatHealthcheck,
			PercentOfAlivedForUp:            oldExtendedServiceData.ServicePercentOfAlivedForUp,
			Type:                            oldExtendedServiceData.ServiceHealthcheckType,
			Timeout:                         oldExtendedServiceData.ServiceHealthcheckTimeout,
			RetriesForUpApplicationServer:   oldExtendedServiceData.ServiceHealthcheckRetriesForUpApplicationServer,
			RetriesForDownApplicationServer: oldExtendedServiceData.ServiceHealthcheckRetriesForDownApplicationServer,
		}

		if oldExtendedServiceData.ServiceExtraInfo != nil {
			currentServiceInfo.ExtraInfo = oldExtendedServiceData.ServiceExtraInfo
		}
		currentServiceInfo.IsUp = oldExtendedServiceData.ServiceIsUp
		tmpBalanceType := oldExtendedServiceData.BalanceType
		tmpRoutingType := oldExtendedServiceData.RoutingType
		tmpProtocol := oldExtendedServiceData.Protocol
		currentServiceInfo.BalanceType = tmpBalanceType
		currentServiceInfo.RoutingType = tmpRoutingType
		currentServiceInfo.Protocol = tmpProtocol

		return nil
	}); err != nil {
		return currentServiceInfo, err
	}

	tmpIncomeServiceIP := incomeServiceData.ServiceIP
	tmpIncomeServicePort := incomeServiceData.ServicePort

	currentServiceInfo.ServiceIP = tmpIncomeServiceIP
	currentServiceInfo.ServicePort = tmpIncomeServicePort
	currentServiceInfo.ApplicationServers = currentApplicationServers

	return currentServiceInfo, nil
}

// LoadAllStorageDataToDomainModels ...
func (storageEntity *StorageEntity) LoadAllStorageDataToDomainModels() ([]*domain.ServiceInfo, error) {
	servicesInfo := []*domain.ServiceInfo{}
	storageEntity.Lock()
	defer storageEntity.Unlock()
	if err := storageEntity.Db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			key := item.Key()
			err := item.Value(func(val []byte) error {
				oldExtendedServiceData := ExtendedServiceData{}
				if err := json.Unmarshal(val, &oldExtendedServiceData); err != nil {
					return fmt.Errorf("can't unmarshall application servers data: %v", err)
				}
				rawServiceData := strings.Split(string(key), ":")
				if len(rawServiceData) != 2 {
					return nil
					// return fmt.Errorf("fail when take service data, expect format x.x.x.x:p, have: %s", key)
				}
				currentApplicationServers := []*domain.ApplicationServer{}
				for _, oldApplicationServer := range oldExtendedServiceData.ApplicationServers {
					renewPointerForOldApplicationServer := oldApplicationServer
					currentApplicationServers = append(currentApplicationServers, &renewPointerForOldApplicationServer)
				}

				hc := domain.ServiceHealthcheck{
					RepeatHealthcheck:               oldExtendedServiceData.ServiceRepeatHealthcheck,
					PercentOfAlivedForUp:            oldExtendedServiceData.ServicePercentOfAlivedForUp,
					Type:                            oldExtendedServiceData.ServiceHealthcheckType,
					Timeout:                         oldExtendedServiceData.ServiceHealthcheckTimeout,
					RetriesForUpApplicationServer:   oldExtendedServiceData.ServiceHealthcheckRetriesForUpApplicationServer,
					RetriesForDownApplicationServer: oldExtendedServiceData.ServiceHealthcheckRetriesForDownApplicationServer,
				}

				serviceInfo := &domain.ServiceInfo{
					ServiceIP:          rawServiceData[0],
					ServicePort:        rawServiceData[1],
					ApplicationServers: currentApplicationServers,
					Healthcheck:        hc,
					ExtraInfo:          oldExtendedServiceData.ServiceExtraInfo,
					IsUp:               oldExtendedServiceData.ServiceIsUp,
					BalanceType:        oldExtendedServiceData.BalanceType,
					RoutingType:        oldExtendedServiceData.RoutingType,
					Protocol:           oldExtendedServiceData.Protocol,
				}
				servicesInfo = append(servicesInfo, serviceInfo)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return servicesInfo, nil
}

// LoadCacheFromStorage ...
func (storageEntity *StorageEntity) LoadCacheFromStorage(oldStorageEntity *StorageEntity) error {
	storageEntity.Lock()
	defer storageEntity.Unlock()
	err := oldStorageEntity.Db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			err := item.Value(func(v []byte) error {
				err := storageEntity.Db.Update(func(txn *badger.Txn) error {
					err := txn.Set(k, v)
					return err
				})
				return err
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// UpdateServiceInfo validate and update service
func (storageEntity *StorageEntity) UpdateServiceInfo(newServiceData *domain.ServiceInfo, eventUUID string) error {
	serviceDataKey, serviceDataValue, err := transformServiceDataForStorageData(newServiceData)
	if err != nil {
		return fmt.Errorf("can't form data for storage: %v", err)
	}
	storageEntity.Lock()
	defer storageEntity.Unlock()
	err = storageEntity.updateDatabaseServiceInfo(serviceDataKey, serviceDataValue)
	if err != nil {
		return fmt.Errorf("can't update storage for new service: %v", err)
	}

	return nil
}

// ReadTunnelInfoForApplicationServer return nil, if key not exist
func (storageEntity *StorageEntity) ReadTunnelInfoForApplicationServer(ip string) *domain.TunnelForApplicationServer {
	dTunnelInfo := &domain.TunnelForApplicationServer{}
	storageEntity.Lock()
	defer storageEntity.Unlock()
	if err := storageEntity.Db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(ip))
		if err != nil {
			return fmt.Errorf("txn.Get fail: %v", err)
		}
		sTunnelInfo := &TunnelForService{}
		if err = item.Value(func(val []byte) error {
			if err := json.Unmarshal(val, &sTunnelInfo); err != nil {
				return fmt.Errorf("can't unmarshall tunnnel data: %v", err)
			}
			return nil
		}); err != nil {
			return err
		}
		if sTunnelInfo.ServicesToTunnelCount == 0 {
			return fmt.Errorf("no tunnel files: %v", sTunnelInfo.ServicesToTunnelCount)
		}

		dTunnelInfo.ApplicationServerIP = ip
		dTunnelInfo.ServicesToTunnelCount = sTunnelInfo.ServicesToTunnelCount
		dTunnelInfo.SysctlConfFile = sTunnelInfo.SysctlConfFile
		dTunnelInfo.TunnelName = sTunnelInfo.TunnelName

		return nil
	}); err != nil {
		return nil
	}
	return dTunnelInfo
}

// UpdateTunnelFilesInfoAtStorage ...
func (storageEntity *StorageEntity) UpdateTunnelFilesInfoAtStorage(tunnelsFilesInfo []*domain.TunnelForApplicationServer) error {
	storageEntity.Lock()
	defer storageEntity.Unlock()
	for _, tunnelFilesInfo := range tunnelsFilesInfo {
		key := []byte(tunnelFilesInfo.ApplicationServerIP)
		transformedTunnelForService := TunnelForService{
			SysctlConfFile:        tunnelFilesInfo.SysctlConfFile,
			TunnelName:            tunnelFilesInfo.TunnelName,
			ServicesToTunnelCount: tunnelFilesInfo.ServicesToTunnelCount,
		}
		tunnelsFilesInfoValue, err := json.Marshal(transformedTunnelForService)
		if err != nil {
			return fmt.Errorf("can't marshal transformedTunnelForService: %v", err)
		}
		if err := updateDb(storageEntity.Db, key, tunnelsFilesInfoValue); err != nil {
			return fmt.Errorf("can't update db: %v", err)
		}
	}
	return nil
}
