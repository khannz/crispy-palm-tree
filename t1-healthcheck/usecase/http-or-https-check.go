package usecase

import (
	"time"

	"github.com/khannz/crispy-palm-tree/lbost1a-healthcheck/domain"
	"github.com/sirupsen/logrus"
)

type HttpOrHttpsCheckEntity struct {
	hcHTTPAndHTTPSWorker domain.HTTPAndHTTPSWorker
	logging              *logrus.Logger
}

func NewHttpOrHttpsCheckEntity(hcHTTPAndHTTPSWorker domain.HTTPAndHTTPSWorker, logging *logrus.Logger) *HttpOrHttpsCheckEntity {
	return &HttpOrHttpsCheckEntity{
		hcHTTPAndHTTPSWorker: hcHTTPAndHTTPSWorker,
		logging:              logging,
	}
}

func (httpOrHttpsCheckEntity *HttpOrHttpsCheckEntity) IsHttpOrHttpsCheckOk(healthcheckAddress string,
	timeout time.Duration,
	fwmark int,
	isHttpCheck bool,
	id string) bool {
	return httpOrHttpsCheckEntity.hcHTTPAndHTTPSWorker.IsHttpOrHttpsCheckOk(healthcheckAddress,
		timeout,
		fwmark,
		isHttpCheck,
		id)
}
