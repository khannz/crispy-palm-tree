package portadapter

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"

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

// NewServiceInfoToStorage add new service to storage
func (storageEntity *StorageEntity) NewServiceInfoToStorage(serviceData *domain.ServiceInfo,
	eventID string) error {
	storageEntity.Lock()
	defer storageEntity.Unlock()
	serviceDataKey := []byte(serviceData.Address)
	serviceDataValue, err := json.Marshal(serviceData)
	if err != nil {
		return fmt.Errorf("can't marshal service data: %v", err)
	}

	err = updateDb(storageEntity.Db, serviceDataKey, serviceDataValue)
	if err != nil {
		return fmt.Errorf("can't update storage for new service: %v", err)
	}

	return nil
}

func updateDb(db *badger.DB, key, value []byte) error {
	return db.Update(func(txn *badger.Txn) error {
		err := txn.Set(key, value)
		return err
	})
}

// RemoveServiceInfoFromStorage ...
func (storageEntity *StorageEntity) RemoveServiceInfoFromStorage(serviceData *domain.ServiceInfo, eventID string) error {
	keyData := []byte(serviceData.Address)
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
func (storageEntity *StorageEntity) GetServiceInfo(serviceData *domain.ServiceInfo, eventID string) (*domain.ServiceInfo, error) {
	storageEntity.Lock()
	defer storageEntity.Unlock()
	dbServiceData := &domain.ServiceInfo{}
	if err := storageEntity.Db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(serviceData.Address))
		if err != nil {
			return fmt.Errorf("txn.Get fail: %v", err)
		}

		if err = item.Value(func(val []byte) error {
			if err := json.Unmarshal(val, &dbServiceData); err != nil {
				return fmt.Errorf("can't unmarshall application servers data: %v", err)
			}
			return nil
		}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return dbServiceData, nil
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
				rawServiceData := strings.Split(string(key), ":")
				if len(rawServiceData) != 2 {
					return nil // it's not service info
				}
				dbServiceData := &domain.ServiceInfo{}
				if err := json.Unmarshal(val, dbServiceData); err != nil {
					return fmt.Errorf("can't unmarshall service data: %v", err)
				}

				servicesInfo = append(servicesInfo, dbServiceData)
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
func (storageEntity *StorageEntity) UpdateServiceInfo(serviceData *domain.ServiceInfo, eventID string) error {
	storageEntity.Lock()
	defer storageEntity.Unlock()

	serviceDataKey := []byte(serviceData.Address)
	serviceDataValue, err := json.Marshal(serviceData)
	if err != nil {
		return fmt.Errorf("can't marshal service data: %v", err)
	}

	err = updateDb(storageEntity.Db, serviceDataKey, serviceDataValue)
	if err != nil {
		return fmt.Errorf("can't update storage for new service: %v", err)
	}

	return nil
}

// ReadTunnelInfoForApplicationServer return nil, if key not exist
func (storageEntity *StorageEntity) ReadTunnelInfoForApplicationServer(ip string) *domain.TunnelForApplicationServer {
	tunnelInfo := &domain.TunnelForApplicationServer{}
	storageEntity.Lock()
	defer storageEntity.Unlock()
	if err := storageEntity.Db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(ip))
		if err != nil {
			return fmt.Errorf("txn.Get fail: %v", err)
		}
		if err = item.Value(func(val []byte) error {
			if err := json.Unmarshal(val, &tunnelInfo); err != nil {
				return fmt.Errorf("can't unmarshall tunnnel data: %v", err)
			}
			return nil
		}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil
	}
	return tunnelInfo
}

// GetAllApplicationServersWhoHaveTunnels bad code :(
func (storageEntity *StorageEntity) GetAllApplicationServersIPWhoHaveTunnels() ([]string, error) {
	var applicationServersIPWithTunnels []string
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
				rawServiceData := strings.Split(string(key), ":")
				if len(rawServiceData) == 1 {
					applicationServersIPWithTunnels = append(applicationServersIPWithTunnels, rawServiceData[0])
					return nil
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return applicationServersIPWithTunnels, err
	}
	return applicationServersIPWithTunnels, nil
}

// RemoveAllTunnelsInfoForApplicationServer
func (storageEntity *StorageEntity) RemoveAllTunnelsInfoForApplicationServer(ip string) error {
	storageEntity.Lock()
	defer storageEntity.Unlock()
	keyData := []byte(ip)

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

// RemoveTunnelsInfoForApplicationServerFromStorage ...
func (storageEntity *StorageEntity) RemoveTunnelsInfoForApplicationServerFromStorage(ip string) error {
	keyData := []byte(ip)
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

// UpdateTunnelFilesInfoAtStorage ...
func (storageEntity *StorageEntity) UpdateTunnelFilesInfoAtStorage(tunnelsFilesInfo []*domain.TunnelForApplicationServer) error {
	storageEntity.Lock()
	defer storageEntity.Unlock()
	for _, tunnelFilesInfo := range tunnelsFilesInfo {
		key := []byte(tunnelFilesInfo.ApplicationServerIP)
		tunnelFileInfoValue, err := json.Marshal(tunnelFilesInfo)
		if err != nil {
			return fmt.Errorf("can't marshal transformedTunnelForService: %v", err)
		}
		if err := updateDb(storageEntity.Db, key, tunnelFileInfoValue); err != nil {
			return fmt.Errorf("can't update db: %v", err)
		}
	}
	return nil
}
