package application

import (
	"time"

	domain "github.com/khannz/crispy-palm-tree/lbost1a-healthcheck/domain"
	"github.com/khannz/crispy-palm-tree/lbost1a-healthcheck/usecase"
	"github.com/sirupsen/logrus"
)

// HCFacade struct
type HCFacade struct {
	HttpAdvancedWorker domain.HTTPAdvancedWorker
	HTTPAndHTTPSWorker domain.HTTPAndHTTPSWorker
	IcmpWorker         domain.ICMPWorker
	TcpWorker          domain.TCPWorker
	IDgenerator        domain.IDgenerator
	Logging            *logrus.Logger
}

// NewHCFacade ...
func NewHCFacade(httpAdvancedWorker domain.HTTPAdvancedWorker,
	httpAndHTTPSWorker domain.HTTPAndHTTPSWorker,
	icmpWorker domain.ICMPWorker,
	tcpWorker domain.TCPWorker,
	idGenerator domain.IDgenerator,
	logging *logrus.Logger) *HCFacade {

	return &HCFacade{
		HttpAdvancedWorker: httpAdvancedWorker,
		HTTPAndHTTPSWorker: httpAndHTTPSWorker,
		IcmpWorker:         icmpWorker,
		TcpWorker:          tcpWorker,
		IDgenerator:        idGenerator,
		Logging:            logging,
	}
}

func (hcFacade *HCFacade) IsHttpAdvancedCheckOk(healthcheckType string,
	healthcheckAddress string,
	nearFieldsMode bool,
	userDefinedData map[string]string,
	timeout time.Duration,
	fwmark int,
	id string) bool {
	newhttpAdvancedCheckEntity := usecase.NewhttpAdvancedCheckEntity(hcFacade.HttpAdvancedWorker, hcFacade.Logging)
	return newhttpAdvancedCheckEntity.IsHttpAdvancedCheckOk(healthcheckType,
		healthcheckAddress,
		nearFieldsMode,
		userDefinedData,
		timeout,
		fwmark,
		id)
}

func (hcFacade *HCFacade) IsHttpOrHttpsCheckOk(healthcheckAddress string,
	uri string,
	validResponseCodes map[int]struct{},
	timeout time.Duration,
	fwmark int,
	isHttpCheck bool,
	id string) bool {
	newHttpOrHttpsCheckEntity := usecase.NewHttpOrHttpsCheckEntity(hcFacade.HTTPAndHTTPSWorker, hcFacade.Logging)
	return newHttpOrHttpsCheckEntity.IsHttpOrHttpsCheckOk(healthcheckAddress,
		uri,
		validResponseCodes,
		timeout,
		fwmark,
		isHttpCheck,
		id)
}

func (hcFacade *HCFacade) IsIcmpCheckOk(ipS string,
	timeout time.Duration,
	fwmark int,
	id string) bool {
	newIcmpCheckEntity := usecase.NewIcmpCheckEntity(hcFacade.IcmpWorker, hcFacade.Logging)
	return newIcmpCheckEntity.IsIcmpCheckOk(ipS,
		timeout,
		fwmark,
		id)
}

func (hcFacade *HCFacade) IsTcpCheckOk(healthcheckAddress string,
	timeout time.Duration,
	fwmark int,
	id string) bool {
	updateServiceEntity := usecase.NewTcpCheckEntity(hcFacade.TcpWorker, hcFacade.Logging)
	return updateServiceEntity.IsTcpCheckOk(healthcheckAddress,
		timeout,
		fwmark,
		id)
}
