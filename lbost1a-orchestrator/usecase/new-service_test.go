package usecase

import (
	"testing"

	"github.com/khannz/crispy-palm-tree/lbost1a-orchestrator/domain"
	"github.com/khannz/crispy-palm-tree/lbost1a-orchestrator/portadapter"
	"github.com/stretchr/testify/assert"
	logger "github.com/thevan4/logrus-wrapper"
)

// TestNewService ...
func TestNewService(t *testing.T) {
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

	createServiceEntityGracefulEnd := NewNewServiceEntity(locker,
		mockCacheDB,
		mockPersistentDB,
		mockTunnelMaker,
		mockHeathcheckEntity,
		mockCommandGenerator,
		gracefulShutdown,
		logging)
	_, errNotNilOne := createServiceEntityGracefulEnd.NewService(currentServiceInfoOne, "")
	assert.NotNil(errNotNilOne)

	gracefulShutdown.ShutdownNow = false
	createServiceEntityOk := NewNewServiceEntity(locker,
		mockCacheDB,
		mockPersistentDB,
		mockTunnelMaker,
		mockHeathcheckEntity,
		mockCommandGenerator,
		gracefulShutdown,
		logging)
	_, errNilOne := createServiceEntityOk.NewService(currentServiceInfoOne, "")
	assert.Nil(errNilOne)
}
