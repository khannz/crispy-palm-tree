package usecase

import (
	"github.com/khannz/crispy-palm-tree/domain"
)

// HeathcheckEntity ...
type HeathcheckEntity struct {
	// locker
	// 	*logger
	// ipvsadm
}

// NewHeathcheckEntity ...
func NewHeathcheckEntity() *HeathcheckEntity {
	return &HeathcheckEntity{}
}

// CheckApplicationServersInServices ...
func (hc *HeathcheckEntity) CheckApplicationServersInServices(serviceInfo []*domain.ServiceInfo) {
	// lock
}
