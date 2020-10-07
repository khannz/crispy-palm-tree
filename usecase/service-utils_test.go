package usecase

import (
	"testing"

	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/khannz/crispy-palm-tree/portadapter"
	"github.com/stretchr/testify/assert"
	logger "github.com/thevan4/logrus-wrapper"
)

// TestValidateRemoveApplicationServers ...
func TestValidateRemoveApplicationServers(t *testing.T) {
	assert := assert.New(t)
	currentApplicattionServers, applicattionServersForRemove, appServer := createApplicationServersForTests()

	errNotNilOne := validateRemoveApplicationServers(currentApplicattionServers, applicattionServersForRemove)
	assert.NotNil(errNotNilOne)

	errNilOne := validateRemoveApplicationServers(currentApplicattionServers, []*domain.ApplicationServer{appServer})
	assert.Nil(errNilOne)

	errNotNilTwo := validateRemoveApplicationServers(currentApplicattionServers, []*domain.ApplicationServer{applicattionServersForRemove[1]})
	assert.NotNil(errNotNilTwo)

	newApplicattionServersForRemove := []*domain.ApplicationServer{applicattionServersForRemove[1], appServer}
	errNotNilThree := validateRemoveApplicationServers(currentApplicattionServers, newApplicattionServersForRemove)
	assert.NotNil(errNotNilThree)

	newApplicattionServersForRemove = append(newApplicattionServersForRemove, applicattionServersForRemove[0])
	errNotNilFour := validateRemoveApplicationServers(currentApplicattionServers, newApplicattionServersForRemove)
	assert.NotNil(errNotNilFour)
}

// TestFormNewApplicationServersSlice ...
func TestFormNewApplicationServersSlice(t *testing.T) {
	currentApplicattionServers, _, appServer := createApplicationServersForTests()

	reNewApplicattionServers := formNewApplicationServersSlice(currentApplicattionServers, []*domain.ApplicationServer{appServer})
	for _, reNewApplicattionServer := range reNewApplicattionServers {
		if reNewApplicattionServer.IP == appServer.IP && reNewApplicattionServer.Port == appServer.Port {
			t.Errorf("application server was not excluded from slice: %v:%v", reNewApplicattionServer.IP, reNewApplicattionServer.Port)
		}
	}
}

// TestForAddApplicationServersFormUpdateServiceInfo
func TestForAddApplicationServersFormUpdateServiceInfo(t *testing.T) {
	currentApplicattionServers, newApplicattionServers, _ := createApplicationServersForTests()
	serviceInfoOne, serviceInfoTwo, _ := createServicesInfoForTests(currentApplicattionServers, newApplicattionServers)
	_, nilErrOne := forAddApplicationServersFormUpdateServiceInfo(serviceInfoOne, serviceInfoTwo, "")
	assert.Nil(t, nilErrOne)
}

// TestForRemoveApplicationServersFormUpdateServiceInfo ...
func TestForRemoveApplicationServersFormUpdateServiceInfo(t *testing.T) {
	currentApplicattionServers, tmpApplicattionServers, appServerForRemove := createApplicationServersForTests()
	serviceInfoOne, _, serviceInfoThree := createServicesInfoForTests(currentApplicattionServers, tmpApplicattionServers)
	serviceInfoThree.ApplicationServers = []*domain.ApplicationServer{appServerForRemove}
	forRemoveApplicationServersFormUpdateServiceInfo(serviceInfoOne, serviceInfoThree, "")
}

// TestCheckRoutingTypeForApplicationServersValid ...
func TestCheckRoutingTypeForApplicationServersValid(t *testing.T) {
	assert := assert.New(t)
	currentApplicattionServers, tmpApplicattionServers, appServerForRemove := createApplicationServersForTests()
	serviceInfoOne, _, serviceInfoThree := createServicesInfoForTests(currentApplicattionServers, tmpApplicattionServers)
	serviceInfoThree.ApplicationServers = []*domain.ApplicationServer{appServerForRemove}
	// newServiceInfo *domain.ServiceInfo, allCurrentServices []*domain.ServiceInfo
	nilErr := checkRoutingTypeForApplicationServersValid(serviceInfoThree, []*domain.ServiceInfo{serviceInfoOne})
	assert.Nil(nilErr)

	serviceInfoThree.RoutingType = "masquerading"
	notNilErr := checkRoutingTypeForApplicationServersValid(serviceInfoThree, []*domain.ServiceInfo{serviceInfoOne})
	assert.NotNil(notNilErr)
}

func TestCheckServiceIPAndPortUnique(t *testing.T) {
	assert := assert.New(t)
	currentApplicattionServers, tmpApplicattionServers, _ := createApplicationServersForTests()
	currentServiceInfoOne, _, _ := createServicesInfoForTests(currentApplicattionServers, tmpApplicattionServers)

	notNilErrOne := checkIPAndPortUnique("111.111.111.111", "111", []*domain.ServiceInfo{currentServiceInfoOne})
	assert.NotNil(notNilErrOne)

	notNilErrTwo := checkIPAndPortUnique("1.1.1.1", "1111", []*domain.ServiceInfo{currentServiceInfoOne})
	assert.NotNil(notNilErrTwo)

	nilErr := checkIPAndPortUnique("9.1.1.1", "9111", []*domain.ServiceInfo{currentServiceInfoOne})
	assert.Nil(nilErr)
}

// TestDecreaseJobs ...
func TestDecreaseJobs(t *testing.T) {
	gs := &domain.GracefulShutdown{
		ShutdownNow:  false,
		UsecasesJobs: 1,
	}
	decreaseJobs(gs)
}

// TestFormTunnelsFilesInfo ...
func TestFormTunnelsFilesInfo(t *testing.T) {
	assert := assert.New(t)
	newLogger := &logger.Logger{
		Output:           []string{"stdout"},
		Level:            "trace",
		Formatter:        "text",
		LogEventLocation: true,
	}
	logging, errCreateLogging := logger.NewLogrusLogger(newLogger)
	assert.Nil(errCreateLogging)

	cacheDB, errCreateCacheDB := portadapter.NewStorageEntity(true, "", logging)
	assert.Nil(errCreateCacheDB)
	testID := "test id for TestFormTunnelsFilesInfo"

	currentApplicattionServers, tmpApplicattionServers, appServer := createApplicationServersForTests()
	currentServiceInfoOne, _, _ := createServicesInfoForTests(currentApplicattionServers, tmpApplicattionServers)

	errPutToDb := cacheDB.NewServiceInfoToStorage(currentServiceInfoOne, testID)
	assert.Nil(errPutToDb)

	FormTunnelsFilesInfo([]*domain.ApplicationServer{appServer}, cacheDB)
}

// TestCheckApplicationServersIPAndPortUnique ...
func TestCheckApplicationServersIPAndPortUnique(t *testing.T) {
	assert := assert.New(t)
	currentApplicattionServers, tmpApplicattionServers, appServer := createApplicationServersForTests()
	currentServiceInfoOne, _, _ := createServicesInfoForTests(currentApplicattionServers, tmpApplicattionServers)

	errNotNilOne := checkApplicationServersIPAndPortUnique([]*domain.ApplicationServer{appServer}, []*domain.ServiceInfo{currentServiceInfoOne})
	assert.NotNil(errNotNilOne)

	appServer.IP = "111.111.111.111"
	appServer.Port = "111"
	errNotNilTwo := checkApplicationServersIPAndPortUnique([]*domain.ApplicationServer{appServer}, []*domain.ServiceInfo{currentServiceInfoOne})
	assert.NotNil(errNotNilTwo)

	appServer.IP = "91.1.1.1"
	appServer.Port = "911"
	errNil := checkApplicationServersIPAndPortUnique([]*domain.ApplicationServer{appServer}, []*domain.ServiceInfo{currentServiceInfoOne})
	assert.Nil(errNil)
}

// TestIsServiceExist ...
func TestIsServiceExist(t *testing.T) {
	assert := assert.New(t)
	currentApplicattionServers, tmpApplicattionServers, _ := createApplicationServersForTests()
	currentServiceInfoOne, _, _ := createServicesInfoForTests(currentApplicattionServers, tmpApplicattionServers)

	notExist := isServiceExist("91.1.1.1", "9111", []*domain.ServiceInfo{currentServiceInfoOne})
	assert.False(notExist)

	exist := isServiceExist("111.111.111.111", "111", []*domain.ServiceInfo{currentServiceInfoOne})
	assert.True(exist)
}

// TestCheckApplicationServersExistInService ...
func TestCheckApplicationServersExistInService(t *testing.T) {
	assert := assert.New(t)
	currentApplicattionServers, tmpApplicattionServers, appServer := createApplicationServersForTests()
	currentServiceInfoOne, _, _ := createServicesInfoForTests(currentApplicattionServers, tmpApplicattionServers)

	errNil := checkApplicationServersExistInService([]*domain.ApplicationServer{appServer}, currentServiceInfoOne)
	assert.Nil(errNil)

	appServer.IP = "91.1.1.1"
	appServer.Port = "911"
	errNotNil := checkApplicationServersExistInService([]*domain.ApplicationServer{appServer}, currentServiceInfoOne)
	assert.NotNil(errNotNil)
}
