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
	HttpsWorker        domain.HTTPSWorker
	IcmpWorker         domain.ICMPWorker
	TcpWorker          domain.TCPWorker
	IDgenerator        domain.IDgenerator
	Logging            *logrus.Logger
}

// NewHCFacade ...
func NewHCFacade(httpAdvancedWorker domain.HTTPAdvancedWorker,
	httpsWorker domain.HTTPSWorker,
	icmpWorker domain.ICMPWorker,
	tcpWorker domain.TCPWorker,
	idGenerator domain.IDgenerator,
	logging *logrus.Logger) *HCFacade {

	return &HCFacade{
		HttpAdvancedWorker: httpAdvancedWorker,
		HttpsWorker:        httpsWorker,
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

func (hcFacade *HCFacade) IsHttpsCheckOk(healthcheckAddress string,
	timeout time.Duration,
	fwmark int,
	id string) bool {
	newhttpsCheckEntity := usecase.NewhttpsCheckEntity(hcFacade.HttpsWorker, hcFacade.Logging)
	return newhttpsCheckEntity.IsHttpsCheckOk(healthcheckAddress,
		timeout,
		fwmark,
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
