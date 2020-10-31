package usecase

import (
	"testing"

	"github.com/khannz/crispy-palm-tree/lbost1a-orchestrator/domain"
	"github.com/khannz/crispy-palm-tree/lbost1a-orchestrator/portadapter"
	"github.com/stretchr/testify/assert"
	logger "github.com/thevan4/logrus-wrapper"
)

// TestRemoveApplicationServers ...
func TestRemoveApplicationServers(t *testing.T) {
	assert := assert.New(t)
	locker := &domain.Locker{}
	mockTunnelMaker := &MockTunnelMaker{}
	mockHeathcheckEntity := &MockHCWorker{}
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
		Address:            "111.111.111.111:111",
		IP:                 "111.111.111.111",
		Port:               "111",
		ApplicationServers: map[string]*domain.ApplicationServer{appSrv.IP + ":" + appSrv.Port: appSrv},
		BalanceType:        "rr",
		RoutingType:        "tunneling",
		Protocol:           "tcp",
	}

	removeApplicationServersEntityGracefulEnd := NewRemoveApplicationServers(locker,
		mockCacheDB,
		mockPersistentDB,
		mockTunnelMaker,
		mockHeathcheckEntity,
		gracefulShutdown,
		logging)
	_, errNotNilOne := removeApplicationServersEntityGracefulEnd.RemoveApplicationServers(serviceInfoForRemoveAppSrvrs, "")
	assert.NotNil(errNotNilOne)

	gracefulShutdown.ShutdownNow = false
	removeApplicationServersEntityOk := NewRemoveApplicationServers(locker,
		mockCacheDB,
		mockPersistentDB,
		mockTunnelMaker,
		mockHeathcheckEntity,
		gracefulShutdown,
		logging)
	errNilPastOne := mockCacheDB.NewServiceInfoToStorage(currentServiceInfoOne, "")
	assert.Nil(errNilPastOne)

	_, errNilOne := removeApplicationServersEntityOk.RemoveApplicationServers(serviceInfoForRemoveAppSrvrs, "")
	assert.Nil(errNilOne)
}
