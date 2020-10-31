package usecase

import (
	"testing"

	"github.com/khannz/crispy-palm-tree/lbost1a-orchestrator/domain"
	"github.com/khannz/crispy-palm-tree/lbost1a-orchestrator/portadapter"
	"github.com/stretchr/testify/assert"
	logger "github.com/thevan4/logrus-wrapper"
)

// TestRemoveService ...
func TestRemoveService(t *testing.T) {
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

	currentApplicattionServers, tmpApplicattionServers, _ := createApplicationServersForTests()
	currentServiceInfoOne, _, _ := createServicesInfoForTests(currentApplicattionServers, tmpApplicattionServers)

	removeServiceGracefulEnd := NewRemoveServiceEntity(locker,
		mockCacheDB,
		mockPersistentDB,
		mockTunnelMaker,
		mockHeathcheckEntity,
		gracefulShutdown,
		logging)
	errNotNilOne := removeServiceGracefulEnd.RemoveService(currentServiceInfoOne, "")
	assert.NotNil(errNotNilOne)

	gracefulShutdown.ShutdownNow = false
	removeServiceFail := NewRemoveServiceEntity(locker,
		mockCacheDB,
		mockPersistentDB,
		mockTunnelMaker,
		mockHeathcheckEntity,
		gracefulShutdown,
		logging)
	errNotNilTwo := removeServiceFail.RemoveService(currentServiceInfoOne, "")
	assert.NotNil(errNotNilTwo)

	errNilPastOne := mockCacheDB.NewServiceInfoToStorage(currentServiceInfoOne, "")
	assert.Nil(errNilPastOne)
	removeServiceOk := NewRemoveServiceEntity(locker,
		mockCacheDB,
		mockPersistentDB,
		mockTunnelMaker,
		mockHeathcheckEntity,
		gracefulShutdown,
		logging)
	errNilOne := removeServiceOk.RemoveService(currentServiceInfoOne, "")
	assert.Nil(errNilOne)
}
