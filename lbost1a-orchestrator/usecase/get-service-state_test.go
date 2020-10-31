package usecase

import (
	"testing"

	"github.com/khannz/crispy-palm-tree/lbost1a-orchestrator/domain"
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
	mockHeathcheckEntity := &MockHCWorker{}

	currentApplicattionServers, tmpApplicattionServers, _ := createApplicationServersForTests()
	currentServiceInfoOne, _, _ := createServicesInfoForTests(currentApplicattionServers, tmpApplicattionServers)

	getServiceStateErr := NewGetServiceStateEntity(locker,
		mockHeathcheckEntity,
		gracefulShutdown,
		logging)
	_, errNotNilOne := getServiceStateErr.GetServiceState(currentServiceInfoOne, "")
	assert.NotNil(errNotNilOne)

	gracefulShutdown.ShutdownNow = false

	getServiceStateOk := NewGetServiceStateEntity(locker,
		mockHeathcheckEntity,
		gracefulShutdown,
		logging)
	serviceMiniInfo := &domain.ServiceInfo{
		Address: "111.111.111.111:111",
		IP:      "111.111.111.111",
		Port:    "111",
	}
	_, errNilOne := getServiceStateOk.GetServiceState(serviceMiniInfo, "")
	assert.Nil(errNilOne)
}
