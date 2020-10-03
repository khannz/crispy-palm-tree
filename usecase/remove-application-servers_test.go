package usecase

import (
	"testing"

	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/khannz/crispy-palm-tree/portadapter"
	"github.com/stretchr/testify/assert"
	logger "github.com/thevan4/logrus-wrapper"
)

// TestRemoveApplicationServers ...
func TestRemoveApplicationServers(t *testing.T) {
	assert := assert.New(t)
	locker := &domain.Locker{}
	mockIPVSWorker := &MockIPVSWorker{}
	mockTunnelMaker := &MockTunnelMaker{}
	mockHeathcheckWorker := &MockHeathcheckWorker{}
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

	currentApplicattionServers, tmpApplicattionServers, appSrv := createApplicationServersForTests()
	currentServiceInfoOne, _, _ := createServicesInfoForTests(currentApplicattionServers, tmpApplicattionServers)

	serviceInfoForRemoveAppSrvrs := &domain.ServiceInfo{
		IP:                 "111.111.111.111",
		Port:               "111",
		ApplicationServers: []*domain.ApplicationServer{appSrv},
		BalanceType:        "rr",
		RoutingType:        "tunneling",
		Protocol:           "tcp",
	}

	removeApplicationServersEntityGracefulEnd := NewRemoveApplicationServers(locker,
		mockIPVSWorker,
		mockCacheDB,
		mockPersistentDB,
		mockTunnelMaker,
		mockHeathcheckWorker,
		gracefulShutdown,
		logging)
	_, errNotNilOne := removeApplicationServersEntityGracefulEnd.RemoveApplicationServers(serviceInfoForRemoveAppSrvrs, "")
	assert.NotNil(errNotNilOne)

	gracefulShutdown.ShutdownNow = false
	removeApplicationServersEntityOk := NewRemoveApplicationServers(locker,
		mockIPVSWorker,
		mockCacheDB,
		mockPersistentDB,
		mockTunnelMaker,
		mockHeathcheckWorker,
		gracefulShutdown,
		logging)
	errNilPastOne := mockCacheDB.NewServiceInfoToStorage(currentServiceInfoOne, "")
	assert.Nil(errNilPastOne)

	_, errNilOne := removeApplicationServersEntityOk.RemoveApplicationServers(serviceInfoForRemoveAppSrvrs, "")
	assert.Nil(errNilOne)
}
