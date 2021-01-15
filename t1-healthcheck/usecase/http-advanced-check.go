package usecase

import (
	"time"

	"github.com/khannz/crispy-palm-tree/lbost1a-healthcheck/domain"
	"github.com/sirupsen/logrus"
)

type HttpAdvancedCheckEntity struct {
	hcHTTPAdvancedWorker domain.HTTPAdvancedWorker
	logging              *logrus.Logger
}

func NewhttpAdvancedCheckEntity(hcHTTPAdvancedWorker domain.HTTPAdvancedWorker, logging *logrus.Logger) *HttpAdvancedCheckEntity {
	return &HttpAdvancedCheckEntity{
		hcHTTPAdvancedWorker: hcHTTPAdvancedWorker,
		logging:              logging,
	}
}

func (httpAdvancedCheckEntity *HttpAdvancedCheckEntity) IsHttpAdvancedCheckOk(healthcheckType string,
	healthcheckAddress string,
	nearFieldsMode bool,
	userDefinedData map[string]string,
	timeout time.Duration,
	fwmark int,
	id string) bool {
	return httpAdvancedCheckEntity.hcHTTPAdvancedWorker.IsHttpAdvancedCheckOk(healthcheckType,
		healthcheckAddress,
		nearFieldsMode,
		userDefinedData,
		timeout,
		fwmark,
		id)
}
