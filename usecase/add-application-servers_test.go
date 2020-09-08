package usecase

import (
	"testing"

	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/khannz/crispy-palm-tree/portadapter"
	"github.com/stretchr/testify/assert"
	logger "github.com/thevan4/logrus-wrapper"
)

// TestAddNewApplicationServers ...
func TestAddNewApplicationServers(t *testing.T) {
	assert := assert.New(t)
	locker := &domain.Locker{}
	mockIPVSWorker := &MockIPVSWorker{}
	mockTunnelMaker := &MockTunnelMaker{}
	mockHeathcheckWorker := &MockHeathcheckWorker{}
	mockCommandGenerator := &MockCommandGenerator{}
	gracefulShutdown := &domain.GracefulShutdown{
		ShutdownNow:  true,
		UsecasesJobs: 0,
	}
	newLogger := &logger.Logger{
		Output:           []string{"stdout"},
		Level:            "trace",
		Formatter:        "text",
		LogEventLocation: true,
	}
	logging, errCreateLogging := logger.NewLogrusLogger(newLogger)
	assert.Nil(errCreateLogging)
	mockCacheDB, errCreateCacheDB := portadapter.NewStorageEntity(true, "", logging)
	assert.Nil(errCreateCacheDB)
	mockPersistentDB, errCreatePersistentDB := portadapter.NewStorageEntity(true, "", logging)
	assert.Nil(errCreatePersistentDB)

	currentApplicattionServers, tmpApplicattionServers, _ := createApplicationServersForTests()
	currentServiceInfoOne, _, _ := createServicesInfoForTests(currentApplicattionServers, tmpApplicattionServers)

	addApplicationServersGracefulEnd := NewAddApplicationServers(locker,
		mockIPVSWorker,
		mockCacheDB,
		mockPersistentDB,
		mockTunnelMaker,
		mockHeathcheckWorker,
		mockCommandGenerator,
		gracefulShutdown,
		logging)
	_, errNotNilOne := addApplicationServersGracefulEnd.AddNewApplicationServers(currentServiceInfoOne, "")
	assert.NotNil(errNotNilOne)

	gracefulShutdown.ShutdownNow = false
	addApplicationServersOk := NewAddApplicationServers(locker,
		mockIPVSWorker,
		mockCacheDB,
		mockPersistentDB,
		mockTunnelMaker,
		mockHeathcheckWorker,
		mockCommandGenerator,
		gracefulShutdown,
		logging)
	_, errNotNilTwo := addApplicationServersOk.AddNewApplicationServers(currentServiceInfoOne, "")
	assert.NotNil(errNotNilTwo)

	errNilPastOne := mockCacheDB.NewServiceInfoToStorage(currentServiceInfoOne, "")
	assert.Nil(errNilPastOne)
	newAppserversInService := &domain.ServiceInfo{
		ServiceIP:          "111.111.111.111",
		ServicePort:        "111",
		ApplicationServers: tmpApplicattionServers,
	}
	_, errNilOne := addApplicationServersOk.AddNewApplicationServers(newAppserversInService, "")
	assert.Nil(errNilOne)
}
