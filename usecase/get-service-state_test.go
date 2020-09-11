package usecase

import (
	"testing"

	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/khannz/crispy-palm-tree/portadapter"
	"github.com/stretchr/testify/assert"
	logger "github.com/thevan4/logrus-wrapper"
)

// TestGetServiceState ...
func TestGetServiceState(t *testing.T) {
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

	getServiceStateErr := NewGetServiceStateEntity(locker,
		mockCacheDB,
		gracefulShutdown,
		logging)
	_, errNotNilOne := getServiceStateErr.GetServiceState(currentServiceInfoOne, "")
	assert.NotNil(errNotNilOne)

	gracefulShutdown.ShutdownNow = false
	_, errNotNilTwo := getServiceStateErr.GetServiceState(currentServiceInfoOne, "")
	assert.NotNil(errNotNilTwo)

	getServiceStateOk := NewGetServiceStateEntity(locker,
		mockCacheDB,
		gracefulShutdown,
		logging)
	serviceMiniInfo := &domain.ServiceInfo{
		ServiceIP:   "111.111.111.111",
		ServicePort: "111",
	}
	_, errNilOne := getServiceStateOk.GetServiceState(serviceMiniInfo, "")
	assert.NotNil(errNilOne)
}