package usecase

import (
	"time"

	"github.com/khannz/crispy-palm-tree/lbost1a-healthcheck/domain"
	"github.com/sirupsen/logrus"
)

type IcmpCheckEntity struct {
	hcICMPWorker domain.ICMPWorker
	logging      *logrus.Logger
}

func NewIcmpCheckEntity(hcICMPWorker domain.ICMPWorker, logging *logrus.Logger) *IcmpCheckEntity {
	return &IcmpCheckEntity{
		hcICMPWorker: hcICMPWorker,
		logging:      logging,
	}
}

func (icmpCheckEntity *IcmpCheckEntity) IsIcmpCheckOk(ipS string,
	seq int,
	timeout time.Duration,
	fwmark int,
	id string) bool {
	return icmpCheckEntity.hcICMPWorker.IsIcmpCheckOk(ipS,
		seq,
		timeout,
		fwmark,
		id)
}
