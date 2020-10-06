package usecase

import (
	"testing"

	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/khannz/crispy-palm-tree/healthchecks"
	"github.com/khannz/crispy-palm-tree/portadapter"
	"github.com/stretchr/testify/assert"
	logger "github.com/thevan4/logrus-wrapper"
)

// TestModifyService ...
func TestModifyService(t *testing.T) {
	assert := assert.New(t)
	locker := &domain.Locker{}
	mockIPVSWorker := &MockIPVSWorker{}
	mockTunnelMaker := &MockTunnelMaker{}
	mockHeathcheckEntity := &healthchecks.HeathcheckEntity{}
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
		mockIPVSWorker,
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
		mockIPVSWorker,
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
		mockIPVSWorker,
		mockCacheDB,
		mockPersistentDB,
		mockTunnelMaker,
		mockHeathcheckEntity,
		mockCommandGenerator,
		gracefulShutdown,
		logging)
	_, errNilOne := modifyServiceOk.ModifyService(currentServiceInfoOne, "")
	assert.Nil(errNilOne)

	tmpAppSrvrs := []*domain.ApplicationServer{}
	tmpAppSrvrs = append(tmpAppSrvrs, currentServiceInfoOne.ApplicationServers...)

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
	serviceForModify.ApplicationServers[0] = tmpApplicattionServers[0]
	notTrueOne := modifyServiceOk.isServicesIPsAndPortsEqual(serviceForModify, currentServiceInfoOne, "")
	assert.False(notTrueOne)

	serviceForModify.ApplicationServers = []*domain.ApplicationServer{}
	notTrueTwo := modifyServiceOk.isServicesIPsAndPortsEqual(serviceForModify, currentServiceInfoOne, "")
	assert.False(notTrueTwo)

	serviceForModify.Port = "9888"
	notTrueThree := modifyServiceOk.isServicesIPsAndPortsEqual(serviceForModify, currentServiceInfoOne, "")
	assert.False(notTrueThree)
}
