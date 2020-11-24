package usecase

import (
	"time"

	"github.com/khannz/crispy-palm-tree/lbost1a-healthcheck/domain"
	"github.com/sirupsen/logrus"
)

type TcpCheckEntity struct {
	hcTCPWorker domain.TCPWorker
	logging     *logrus.Logger
}

func NewTcpCheckEntity(hcTCPWorker domain.TCPWorker, logging *logrus.Logger) *TcpCheckEntity {
	return &TcpCheckEntity{
		hcTCPWorker: hcTCPWorker,
		logging:     logging,
	}
}

func (tcpCheckEntity *TcpCheckEntity) IsTcpCheckOk(healthcheckAddress string,
	timeout time.Duration,
	fwmark int,
	id string) bool {
	return tcpCheckEntity.hcTCPWorker.IsTcpCheckOk(healthcheckAddress,
		timeout,
		fwmark,
		id)
}
