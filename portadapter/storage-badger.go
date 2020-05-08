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
	ServiceHealthcheckType    string                     `json:"serviceHealthcheckType"`
	ServiceHealthcheckTimeout time.Duration              `json:"serviceHealthcheckTimeout"`
	ServiceExtraInfo          []string                   `json:"serviceExtraInfo"`
	ServiceState              bool                       `json:"serviceState"`
	ApplicationServers        []domain.ApplicationServer `json:"applicationServers"`
}

// NewServiceDataToStorage add new service to storage. Also check unique data
func (storageEntity *StorageEntity) NewServiceDataToStorage(serviceData *domain.ServiceInfo,
	eventUUID string) error {
	serviceDataKey, serviceDataValue, err := transformServiceDataForStorageData(serviceData)
	if err != nil {
		return fmt.Errorf("can't form data for storage: %v", err)
	}

	err = storageEntity.checkUnique(serviceDataKey, serviceData)
	if err != nil {
		return fmt.Errorf("some key not unique: %v", err)
	}

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
		ServiceHealthcheckType:    serviceData.Healthcheck.Type,
		ServiceHealthcheckTimeout: serviceData.Healthcheck.Timeout,
		ServiceExtraInfo:          serviceData.ExtraInfo,
		ServiceState:              serviceData.State,
		ApplicationServers:        renewApplicationServers,
	}
	serviceDataValue, err := json.Marshal(transformedServiceData)
	if err != nil {
		return nil, nil, fmt.Errorf("can't marshal transformedServiceData: %v", err)
	}
	return serviceDataKey, serviceDataValue, nil
}

// func transformStorageDataFormServiceData(serviceDataKey, serviceDataValue []byte) *domain.ServiceInfo {
// 	serviceDataKey
// }

func updateDb(db *badger.DB, key, value []byte) error {
	return db.Update(func(txn *badger.Txn) error {
		err := txn.Set(key, value)
		return err
	})
}

func (storageEntity *StorageEntity) checkServiceUniqueInDatabase(key []byte) error {
	return storageEntity.Db.View(func(txn *badger.Txn) error {
		_, err := txn.Get(key)
		if err != nil {
			return nil
		}
		return fmt.Errorf("key %s already exist in storage", key)
	})
}

func (storageEntity *StorageEntity) checkUnique(serviceDataKey []byte,
	serviceInfo *domain.ServiceInfo) error {
	var err error
	err = storageEntity.checkServiceUniqueInDatabase(serviceDataKey)
	if err != nil {
		return fmt.Errorf("service is not unique: %v", err)
	}
	err = storageEntity.checkUniqueInApplicationServersFromDatabase(serviceInfo.ServiceIP, serviceInfo.ServicePort)
	if err != nil {
		return fmt.Errorf("service is not unique, find it in applicatation servers: %v", err)
	}

	for _, applicationServer := range serviceInfo.ApplicationServers {
		err = storageEntity.checkUniqueInApplicationServersFromDatabase(applicationServer.ServerIP, applicationServer.ServerPort)
		if err != nil {
			return fmt.Errorf("application server for service %s is not unique: %v", serviceDataKey, err)
		}
	}
	return nil
}

// TODO: need better check unique, app srv to services too
func (storageEntity *StorageEntity) checkUniqueInApplicationServersFromDatabase(checkIP,
	checkPort string) error {
	if err := storageEntity.Db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			serviceData := item.Key()
			err := item.Value(func(v []byte) error {
				oldExtendedServiceData := ExtendedServiceData{}
				if err := json.Unmarshal(v, &oldExtendedServiceData); err != nil {
					return fmt.Errorf("can't unmarshall application servers data: %v", err)
				}

				for _, oldApplicationServer := range oldExtendedServiceData.ApplicationServers {
					if checkIP == oldApplicationServer.ServerIP &&
						checkPort == oldApplicationServer.ServerPort {
						return fmt.Errorf("in service %s application server %v already exist", serviceData, oldApplicationServer)
					}
				}

				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (storageEntity *StorageEntity) updateDatabaseServiceInfo(serviceDataKey,
	serviceDataValue []byte) error {
	var err error
	err = updateDb(storageEntity.Db, serviceDataKey, serviceDataValue)
	if err != nil {
		return fmt.Errorf("can't update db for service data: %v", err)
	}
	return nil
}

// RemoveServiceDataFromStorage ...
func (storageEntity *StorageEntity) RemoveServiceDataFromStorage(serviceData *domain.ServiceInfo, eventUUID string) error {
	keyData := []byte(serviceData.ServiceIP + ":" + serviceData.ServicePort)
	if err := storageEntity.checkServiceUniqueInDatabase(keyData); err == nil { // if err not nil key exist
		return fmt.Errorf("key %s not exist in database", keyData)
	}

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
		Type:    "",
		Timeout: time.Duration(999 * time.Second),
	}
	currentServiceInfo := &domain.ServiceInfo{
		ServiceIP:          "",
		ServicePort:        "",
		ApplicationServers: []*domain.ApplicationServer{},
		Healthcheck:        shc,
		ExtraInfo:          []string{},
		State:              false,
	}
	currentApplicationServers := []*domain.ApplicationServer{}
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

		hc := domain.ServiceHealthcheck{
			Type:    "",
			Timeout: time.Duration(999 * time.Second),
		}
		if oldExtendedServiceData.ServiceHealthcheckType != "" {
			hc.Type = oldExtendedServiceData.ServiceHealthcheckType
		}
		if oldExtendedServiceData.ServiceHealthcheckTimeout != time.Duration(999*time.Second) {
			hc.Timeout = oldExtendedServiceData.ServiceHealthcheckTimeout
		}
		currentServiceInfo.Healthcheck = hc

		if oldExtendedServiceData.ServiceExtraInfo != nil {
			currentServiceInfo.ExtraInfo = oldExtendedServiceData.ServiceExtraInfo
		}
		if oldExtendedServiceData.ServiceState {
			currentServiceInfo.State = oldExtendedServiceData.ServiceState
		}

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

// LoadAllStorageDataToDomainModel ...
func (storageEntity *StorageEntity) LoadAllStorageDataToDomainModel() ([]*domain.ServiceInfo, error) {
	servicesInfo := []*domain.ServiceInfo{}

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
					return fmt.Errorf("fail when take service data, expect format x.x.x.x:p, have: %s", key)
				}
				currentApplicationServers := []*domain.ApplicationServer{}
				for _, oldApplicationServer := range oldExtendedServiceData.ApplicationServers {
					renewPointerForOldApplicationServer := oldApplicationServer
					currentApplicationServers = append(currentApplicationServers, &renewPointerForOldApplicationServer)
				}

				hc := domain.ServiceHealthcheck{
					Type:    "",
					Timeout: time.Duration(999 * time.Second),
				}
				if oldExtendedServiceData.ServiceHealthcheckType != "" {
					hc.Type = oldExtendedServiceData.ServiceHealthcheckType
				}
				if oldExtendedServiceData.ServiceHealthcheckTimeout != time.Duration(999*time.Second) {
					hc.Timeout = oldExtendedServiceData.ServiceHealthcheckTimeout
				}

				serviceInfo := &domain.ServiceInfo{
					ServiceIP:          rawServiceData[0],
					ServicePort:        rawServiceData[1],
					ApplicationServers: currentApplicationServers,
					Healthcheck:        hc,
					ExtraInfo:          oldExtendedServiceData.ServiceExtraInfo,
					State:              oldExtendedServiceData.ServiceState,
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

	err = storageEntity.updateDatabaseServiceInfo(serviceDataKey, serviceDataValue)
	if err != nil {
		return fmt.Errorf("can't update storage for new service: %v", err)
	}

	return nil
}
