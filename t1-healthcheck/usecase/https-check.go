package usecase

import (
	"time"

	"github.com/khannz/crispy-palm-tree/lbost1a-healthcheck/domain"
	"github.com/sirupsen/logrus"
)

type HttpsCheckEntity struct {
	hcHTTPSWorker domain.HTTPSWorker
	logging       *logrus.Logger
}

func NewhttpsCheckEntity(hcHTTPSWorker domain.HTTPSWorker, logging *logrus.Logger) *HttpsCheckEntity {
	return &HttpsCheckEntity{
		hcHTTPSWorker: hcHTTPSWorker,
		logging:       logging,
	}
}

func (httpsCheckEntity *HttpsCheckEntity) IsHttpsCheckOk(healthcheckAddress string,
	timeout time.Duration,
	fwmark int,
	id string) bool {
	return httpsCheckEntity.hcHTTPSWorker.IsHttpsCheckOk(healthcheckAddress,
		timeout,
		fwmark,
		id)
}
