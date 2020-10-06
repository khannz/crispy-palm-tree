package usecase

import (
	"testing"

	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/khannz/crispy-palm-tree/healthcheck"
	"github.com/khannz/crispy-palm-tree/portadapter"
	"github.com/stretchr/testify/assert"
	logger "github.com/thevan4/logrus-wrapper"
)

// TestAddNewApplicationServers ...
func TestAddNewApplicationServers(t *testing.T) {
	assert := assert.New(t)
	locker := &domain.Locker{}
	mockTunnelMaker := &MockTunnelMaker{}
	mockHeathcheckEntity := &healthcheck.HeathcheckEntity{}
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
		mockCacheDB,
		mockPersistentDB,
		mockTunnelMaker,
		mockHeathcheckEntity,
		mockCommandGenerator,
		gracefulShutdown,
		logging)
	_, errNotNilOne := addApplicationServersGracefulEnd.AddNewApplicationServers(currentServiceInfoOne, "")
	assert.NotNil(errNotNilOne)

	gracefulShutdown.ShutdownNow = false
	addApplicationServersOk := NewAddApplicationServers(locker,
		mockCacheDB,
		mockPersistentDB,
		mockTunnelMaker,
		mockHeathcheckEntity,
		mockCommandGenerator,
		gracefulShutdown,
		logging)
	_, errNotNilTwo := addApplicationServersOk.AddNewApplicationServers(currentServiceInfoOne, "")
	assert.NotNil(errNotNilTwo)

	errNilPastOne := mockCacheDB.NewServiceInfoToStorage(currentServiceInfoOne, "")
	assert.Nil(errNilPastOne)
	newAppserversInService := &domain.ServiceInfo{
		Address:            "111.111.111.111:111",
		IP:                 "111.111.111.111",
		Port:               "111",
		ApplicationServers: tmpApplicattionServers,
	}
	_, errNilOne := addApplicationServersOk.AddNewApplicationServers(newAppserversInService, "")
	assert.Nil(errNilOne)
}
