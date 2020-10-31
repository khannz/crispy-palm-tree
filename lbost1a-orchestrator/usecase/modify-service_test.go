package usecase

import (
	"testing"

	"github.com/khannz/crispy-palm-tree/lbost1a-orchestrator/domain"
	"github.com/khannz/crispy-palm-tree/lbost1a-orchestrator/portadapter"
	"github.com/stretchr/testify/assert"
	logger "github.com/thevan4/logrus-wrapper"
)

// TestModifyService ...
func TestModifyService(t *testing.T) {
	assert := assert.New(t)
	locker := &domain.Locker{}
	mockTunnelMaker := &MockTunnelMaker{}
	mockHeathcheckEntity := &MockHCWorker{}
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

	modifyServiceGracefulEnd := NewModifyServiceEntity(locker,
		mockCacheDB,
		mockPersistentDB,
		mockTunnelMaker,
		mockHeathcheckEntity,
		mockCommandGenerator,
		gracefulShutdown,
		logging)
	_, errNotNilOne := modifyServiceGracefulEnd.ModifyService(currentServiceInfoOne, "")
	assert.NotNil(errNotNilOne)

	gracefulShutdown.ShutdownNow = false
	modifyServiceFail := NewModifyServiceEntity(locker,
		mockCacheDB,
		mockPersistentDB,
		mockTunnelMaker,
		mockHeathcheckEntity,
		mockCommandGenerator,
		gracefulShutdown,
		logging)
	_, errNotNilTwo := modifyServiceFail.ModifyService(currentServiceInfoOne, "")
	assert.NotNil(errNotNilTwo)

	errNilPastOne := mockCacheDB.NewServiceInfoToStorage(currentServiceInfoOne, "")
	assert.Nil(errNilPastOne)
	currentServiceInfoOne.BalanceType = "wrr"
	modifyServiceOk := NewModifyServiceEntity(locker,
		mockCacheDB,
		mockPersistentDB,
		mockTunnelMaker,
		mockHeathcheckEntity,
		mockCommandGenerator,
		gracefulShutdown,
		logging)
	_, errNilOne := modifyServiceOk.ModifyService(currentServiceInfoOne, "")
	assert.Nil(errNilOne)

	tmpAppSrvrs := make(map[string]*domain.ApplicationServer)
	for k, v := range currentServiceInfoOne.ApplicationServers {
		tmpAppSrvrs[k] = v
	}

	serviceForModify := &domain.ServiceInfo{
		Address:            currentServiceInfoOne.Address,
		IP:                 currentServiceInfoOne.IP,
		Port:               currentServiceInfoOne.Port,
		ApplicationServers: tmpAppSrvrs,
		// Healthcheck:        currentServiceInfoOne.Healthcheck,
		IsUp:        currentServiceInfoOne.IsUp,
		BalanceType: currentServiceInfoOne.BalanceType,
		RoutingType: currentServiceInfoOne.RoutingType,
		Protocol:    currentServiceInfoOne.Protocol,
	}

	serviceForModify.ApplicationServers["3.3.3.3"+":"+"3333"] = tmpApplicattionServers["3.3.3.3"+":"+"3333"]
	notTrueOne := modifyServiceOk.isServicesIPsAndPortsEqual(serviceForModify, currentServiceInfoOne, "")
	assert.False(notTrueOne)

	serviceForModify.ApplicationServers = map[string]*domain.ApplicationServer{}
	notTrueTwo := modifyServiceOk.isServicesIPsAndPortsEqual(serviceForModify, currentServiceInfoOne, "")
	assert.False(notTrueTwo)

	serviceForModify.Port = "9888"
	notTrueThree := modifyServiceOk.isServicesIPsAndPortsEqual(serviceForModify, currentServiceInfoOne, "")
	assert.False(notTrueThree)
}
