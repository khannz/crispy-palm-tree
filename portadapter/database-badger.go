package portadapter

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"

	badger "github.com/dgraph-io/badger/v2"
	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/sirupsen/logrus"
)

// DatabaseEntity ...
type DatabaseEntity struct {
	sync.Mutex
	db      *badger.DB
	logging *logrus.Logger
}

// NewDatabaseEntity ...
// do not forget defer db.Close()!
func NewDatabaseEntity(isInMemory bool, dbPath string, logging *logrus.Logger) (*DatabaseEntity, error) {
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
		return nil, fmt.Errorf("can't open database: %v", err)
	}
	databaseEntity := &DatabaseEntity{
		db:      db,
		logging: logging,
	}
	return databaseEntity, nil
}

func optionsForDbInMemory(logger *logrus.Logger) badger.Options {
	defOpt := badger.DefaultOptions("")
	defOpt.Logger = logger
	defOpt.InMemory = true
	return defOpt
}

func optionsForDbPersistent(dbPath string, logger *logrus.Logger) badger.Options {
	defOpt := badger.DefaultOptions("dbPath")
	defOpt.Logger = logger
	return defOpt
}

// ServiceDataToDatabase add new service to database. Also check unique data
func ServiceDataToDatabase(db *badger.DB, serviceData domain.ServiceInfo) error {
	serviceDataKey, serviceDataValue, applicationServersData, err := formDataForDatabase(serviceData)
	if err != nil {
		return fmt.Errorf("can't form data for database: %v", err)
	}

	err = checkUnique(db, serviceDataKey, applicationServersData)
	if err != nil {
		return fmt.Errorf("some key not unique: %v", err)
	}

	err = updateDatabaseForNewservice(db, serviceDataKey, serviceDataValue, applicationServersData)
	if err != nil {
		return fmt.Errorf("can't update database for new service: %v", err)
	}

	return nil
}

func transformServiceDataForDatabase(serviceData domain.ServiceInfo) ([]byte,
	[]byte,
	error) {
	serviceDataKey, err := json.Marshal(serviceData.ServiceIP + "-" + serviceData.ServicePort)
	if err != nil {
		return nil, nil, fmt.Errorf("can't marshal serviceData: %v", err)
	}
	serviceDataValue, err := json.Marshal(serviceData.ApplicationServers)
	if err != nil {
		return nil, nil, fmt.Errorf("can't marshal rawServiceDataValue: %v", err)
	}
	return serviceDataKey, serviceDataValue, nil
}

func transformApplicationServersForDatabase(serviceData domain.ServiceInfo) ([][]byte, error) {
	applicationServersSlice := [][]byte{}
	for _, applicationServer := range serviceData.ApplicationServers {
		rawApplicationServerData := applicationServer.ServerIP + ":" + applicationServer.ServerPort
		applicationServerData, err := json.Marshal(rawApplicationServerData)
		if err != nil {
			return applicationServersSlice, fmt.Errorf("can't marshal application server data %v: %v", rawApplicationServerData, err)
		}
		applicationServersSlice = append(applicationServersSlice, applicationServerData)
	}
	return applicationServersSlice, nil
}

func updateDb(db *badger.DB, key, value []byte) error {
	return db.Update(func(txn *badger.Txn) error {
		err := txn.Set(key, value)
		return err
	})
}

func checkUniqueInDatabase(db *badger.DB, key []byte) error {
	return db.View(func(txn *badger.Txn) error {
		_, err := txn.Get(key)
		if err == nil {
			return fmt.Errorf("key %s already exist in database", key)
		}
		return nil
	})
}

func formDataForDatabase(serviceData domain.ServiceInfo) ([]byte,
	[]byte,
	[][]byte,
	error) {
	serviceDataKey, serviceDataValue, err := transformServiceDataForDatabase(serviceData)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("can't transform service data for database: %v", err)
	}

	applicationServersData, err := transformApplicationServersForDatabase(serviceData)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("can't transform application servers for database: %v", err)
	}
	return serviceDataKey, serviceDataValue, applicationServersData, nil
}

func checkUnique(db *badger.DB,
	serviceDataKey []byte,
	applicationServersData [][]byte) error {
	var err error
	err = checkUniqueInDatabase(db, serviceDataKey)
	if err != nil {
		return fmt.Errorf("service is not unique: %v", err)
	}

	for _, applicationServerData := range applicationServersData {
		err = checkUniqueInDatabase(db, applicationServerData)
		if err != nil {
			return fmt.Errorf("application server is not unique: %v", err)
		}
	}
	return nil
}

func updateDatabaseForNewservice(db *badger.DB,
	serviceDataKey,
	serviceDataValue []byte,
	applicationServersData [][]byte) error {
	var err error
	err = updateDb(db, serviceDataKey, serviceDataValue)
	if err != nil {
		return fmt.Errorf("can't update db for service data: %v", err)
	}

	for _, applicationServerData := range applicationServersData {
		err = updateDb(db, applicationServerData, []byte(""))
		if err != nil {
			return fmt.Errorf("can't update db for application server data: %v", err)
		}
	}
	return nil
}
