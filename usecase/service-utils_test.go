package usecase

import (
	"testing"

	"github.com/khannz/crispy-palm-tree/domain"
	"github.com/stretchr/testify/assert"
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
		if reNewApplicattionServer.ServerIP == appServer.ServerIP && reNewApplicattionServer.ServerPort == appServer.ServerPort {
			t.Errorf("application server was not excluded from slice: %v:%v", reNewApplicattionServer.ServerIP, reNewApplicattionServer.ServerPort)
		}
	}
}

// TestForAddApplicationServersFormUpdateServiceInfo
func TestForAddApplicationServersFormUpdateServiceInfo(t *testing.T) {

}

func createApplicationServersForTests() ([]*domain.ApplicationServer, []*domain.ApplicationServer, *domain.ApplicationServer) {
	serverHealthcheckOne, serverHealthcheckTwo, serverHealthcheckThree, serverHealthcheckFour, serverHealthcheckFive := createServerHealthchecksForTests()
	appSrvOne := &domain.ApplicationServer{
		ServerIP:          "1.1.1.1",
		ServerPort:        "1111",
		IsUp:              true,
		ServerHealthcheck: serverHealthcheckOne,
	}
	appSrvTwo := &domain.ApplicationServer{
		ServerIP:          "2.2.2.2",
		ServerPort:        "2222",
		IsUp:              true,
		ServerHealthcheck: serverHealthcheckTwo,
	}
	appSrvThree := &domain.ApplicationServer{
		ServerIP:          "3.3.3.3",
		ServerPort:        "3333",
		IsUp:              true,
		ServerHealthcheck: serverHealthcheckThree,
	}
	appSrvFour := &domain.ApplicationServer{
		ServerIP:          "4.4.4.4",
		ServerPort:        "4444",
		IsUp:              true,
		ServerHealthcheck: serverHealthcheckFour,
	}
	appSrvFive := &domain.ApplicationServer{
		ServerIP:          "5.5.5.5",
		ServerPort:        "5555",
		IsUp:              true,
		ServerHealthcheck: serverHealthcheckFive,
	}
	appSrvSix := &domain.ApplicationServer{
		ServerIP:          "1.1.1.1",
		ServerPort:        "1111",
		IsUp:              true,
		ServerHealthcheck: serverHealthcheckOne,
	}
	return []*domain.ApplicationServer{appSrvOne, appSrvTwo, appSrvFive}, []*domain.ApplicationServer{appSrvThree, appSrvFour}, appSrvSix
}

func createServerHealthchecksForTests() (domain.ServerHealthcheck, domain.ServerHealthcheck, domain.ServerHealthcheck, domain.ServerHealthcheck, domain.ServerHealthcheck) {
	advancedHealthcheckParametersOne, advancedHealthcheckParametersTwo, advancedHealthcheckParametersThree, advancedHealthcheckParametersFour := createAdvancedHealthcheckParametersForTests()

	serverHealthcheckOne := domain.ServerHealthcheck{
		TypeOfCheck:                   "http-advanced",
		HealthcheckAddress:            "127.0.1.1:9000",
		AdvancedHealthcheckParameters: []domain.AdvancedHealthcheckParameters{advancedHealthcheckParametersOne, advancedHealthcheckParametersTwo},
	}

	serverHealthcheckTwo := domain.ServerHealthcheck{
		TypeOfCheck:        "icmp",
		HealthcheckAddress: "127.0.2.1",
	}

	serverHealthcheckThree := domain.ServerHealthcheck{
		TypeOfCheck:                   "http-advanced",
		HealthcheckAddress:            "127.0.3.1:8000",
		AdvancedHealthcheckParameters: []domain.AdvancedHealthcheckParameters{advancedHealthcheckParametersThree, advancedHealthcheckParametersFour},
	}

	serverHealthcheckFour := domain.ServerHealthcheck{
		TypeOfCheck:        "tcp",
		HealthcheckAddress: "127.0.4.1",
	}

	serverHealthcheckFive := domain.ServerHealthcheck{
		TypeOfCheck:        "tcp",
		HealthcheckAddress: "127.0.5.1",
	}

	return serverHealthcheckOne, serverHealthcheckTwo, serverHealthcheckThree, serverHealthcheckFour, serverHealthcheckFive
}

func createAdvancedHealthcheckParametersForTests() (domain.AdvancedHealthcheckParameters, domain.AdvancedHealthcheckParameters, domain.AdvancedHealthcheckParameters, domain.AdvancedHealthcheckParameters) {
	advancedHealthcheckParametersOne := domain.AdvancedHealthcheckParameters{
		NearFieldsMode:  true,
		UserDefinedData: map[string]interface{}{"one": "oneValue", "two": 2},
	}
	advancedHealthcheckParametersTwo := domain.AdvancedHealthcheckParameters{
		NearFieldsMode:  false,
		UserDefinedData: map[string]interface{}{"three": "threeValue", "four": 4},
	}

	advancedHealthcheckParametersThree := domain.AdvancedHealthcheckParameters{
		NearFieldsMode:  true,
		UserDefinedData: map[string]interface{}{"five": "fiveValue", "six": 6},
	}
	advancedHealthcheckParametersFour := domain.AdvancedHealthcheckParameters{
		NearFieldsMode:  false,
		UserDefinedData: map[string]interface{}{"seven": "sevenValue", "eight": 8},
	}
	return advancedHealthcheckParametersOne, advancedHealthcheckParametersTwo, advancedHealthcheckParametersThree, advancedHealthcheckParametersFour
}

func Test_validateRemoveApplicationServers(t *testing.T) {
	type args struct {
		currentApplicattionServers   []*domain.ApplicationServer
		applicattionServersForRemove []*domain.ApplicationServer
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateRemoveApplicationServers(tt.args.currentApplicattionServers, tt.args.applicattionServersForRemove); (err != nil) != tt.wantErr {
				t.Errorf("validateRemoveApplicationServers() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
