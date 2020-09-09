package usecase

import (
	"testing"

	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/khannz/crispy-palm-tree/portadapter"
	"github.com/stretchr/testify/assert"
	logger "github.com/thevan4/logrus-wrapper"
)

// TestGetAllServices ...
func TestGetAllServices(t *testing.T) {
	assert := assert.New(t)
	locker := &domain.Locker{}
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

	currentApplicattionServers, tmpApplicattionServers, _ := createApplicationServersForTests()
	currentServiceInfoOne, _, _ := createServicesInfoForTests(currentApplicattionServers, tmpApplicattionServers)

	getAllServicesErr := NewGetAllServices(locker,
		mockCacheDB,
		gracefulShutdown,
		logging)
	_, errNotNilOne := getAllServicesErr.GetAllServices("")
	assert.NotNil(errNotNilOne)

	gracefulShutdown.ShutdownNow = false
	_, errNotNilTwo := getAllServicesErr.GetAllServices("")
	assert.Nil(errNotNilTwo)

	errNilPastOne := mockCacheDB.NewServiceInfoToStorage(currentServiceInfoOne, "")
	assert.Nil(errNilPastOne)
	getServiceStateOk := NewGetAllServices(locker,
		mockCacheDB,
		gracefulShutdown,
		logging)

	_, errNilOne := getServiceStateOk.GetAllServices("")
	assert.Nil(errNilOne)
}
