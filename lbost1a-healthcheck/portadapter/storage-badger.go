package portadapter

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"

	badger "github.com/dgraph-io/badger/v2"
	domain "github.com/khannz/crispy-palm-tree/lbost1a-healthcheck/domain"
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

// NewHCServiceToStorage add new service to storage
func (storageEntity *StorageEntity) NewHCServiceToStorage(serviceData *domain.HCService,
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

// RemoveHCServiceFromStorage ...
func (storageEntity *StorageEntity) RemoveHCServiceFromStorage(serviceData *domain.HCService, eventID string) error {
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

// GetHCService ...
func (storageEntity *StorageEntity) GetHCService(serviceData *domain.HCService, eventID string) (*domain.HCService, error) {
	storageEntity.Lock()
	defer storageEntity.Unlock()
	dbServiceData := &domain.HCService{}
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
func (storageEntity *StorageEntity) LoadAllStorageDataToDomainModels() ([]*domain.HCService, error) {
	servicesInfo := []*domain.HCService{}
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
				dbServiceData := &domain.HCService{}
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

// UpdateHCService validate and update service
func (storageEntity *StorageEntity) UpdateHCService(serviceData *domain.HCService, eventID string) error {
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
