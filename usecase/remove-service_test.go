package usecase

import (
	"testing"

	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/khannz/crispy-palm-tree/portadapter"
	"github.com/stretchr/testify/assert"
	logger "github.com/thevan4/logrus-wrapper"
)

// TestRemoveService ...
func TestRemoveService(t *testing.T) {
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

	removeServiceGracefulEnd := NewModifyServiceEntity(locker,
		mockIPVSWorker,
		mockCacheDB,
		mockPersistentDB,
		mockTunnelMaker,
		mockHeathcheckWorker,
		mockCommandGenerator,
		gracefulShutdown,
		logging)
	_, errNotNilOne := removeServiceGracefulEnd.ModifyService(currentServiceInfoOne, "")
	assert.NotNil(errNotNilOne)

	gracefulShutdown.ShutdownNow = false
	removeServiceFail := NewRemoveServiceEntity(locker,
		mockIPVSWorker,
		mockCacheDB,
		mockPersistentDB,
		mockTunnelMaker,
		mockHeathcheckWorker,
		gracefulShutdown,
		logging)
	errNotNilTwo := removeServiceFail.RemoveService(currentServiceInfoOne, "")
	assert.NotNil(errNotNilTwo)

	errNilPastOne := mockCacheDB.NewServiceInfoToStorage(currentServiceInfoOne, "")
	assert.Nil(errNilPastOne)
	removeServiceOk := NewRemoveServiceEntity(locker,
		mockIPVSWorker,
		mockCacheDB,
		mockPersistentDB,
		mockTunnelMaker,
		mockHeathcheckWorker,
		gracefulShutdown,
		logging)
	errNilOne := removeServiceOk.RemoveService(currentServiceInfoOne, "")
	assert.Nil(errNilOne)
}
