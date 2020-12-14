package portadapter

import (
	"net"
	"time"

	"github.com/sirupsen/logrus"
)

const tcpHealthcheckName = "tcp healthcheck"

type TcpEntity struct {
	logging *logrus.Logger
}

func NewTcpEntity(logging *logrus.Logger) *TcpEntity {
	return &TcpEntity{logging: logging}
}

func (tcpEntity *TcpEntity) IsTcpCheckOk(healthcheckAddress string,
	timeout time.Duration,
	fwmark int,
	id string) bool {
	ip, port, err := net.SplitHostPort(healthcheckAddress)
	if err != nil {
		tcpEntity.logging.WithFields(logrus.Fields{
			"entity":   tcpHealthcheckName,
			"event id": id,
		}).Errorf("tcp can't SplitHostPort from %v: %v", healthcheckAddress, err)
		return false
	}

	sockClose := make(chan struct{}, 1)
	defer func(sockClose chan struct{}) { close(sockClose) }(sockClose)
	defer func(sockClose chan struct{}) { sockClose <- struct{}{} }(sockClose)
	tcpConn, err := dialTCP("tcp4", ip, port, timeout, fwmark, sockClose)
	if err != nil {
		tcpEntity.logging.WithFields(logrus.Fields{
			"entity":   tcpHealthcheckName,
			"event id": id,
		}).Tracef("tcp connect to %v error: %v", healthcheckAddress, err)
		return false
	}
	defer tcpConn.f.Close()
	conn := net.Conn(tcpConn.netConn)
	defer conn.Close()

	if conn != nil {
		tcpEntity.logging.WithFields(logrus.Fields{
			"entity":   tcpHealthcheckName,
			"event id": id,
		}).Tracef("tcp connect to %v ok", healthcheckAddress)
		return true
	}

	// somehow it can be..
	tcpEntity.logging.WithFields(logrus.Fields{
		"entity":   tcpHealthcheckName,
		"event id": id,
	}).Error("tcp connect unknown error: connection is nil, but have no errors")
	return false
}
