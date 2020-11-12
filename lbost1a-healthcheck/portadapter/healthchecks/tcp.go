package healthchecks

import (
	"net"
	"time"

	"github.com/sirupsen/logrus"
)

const tcpHealthcheckName = "tcp healthcheck"

func TcpCheckOk(healthcheckAddress string,
	timeout time.Duration,
	fwmark int,
	id string,
	logging *logrus.Logger) bool {
	ip, port, err := net.SplitHostPort(healthcheckAddress)
	if err != nil {
		logging.WithFields(logrus.Fields{
			"entity":   tcpHealthcheckName,
			"event id": id,
		}).Errorf("tcp can't SplitHostPort from %v: %v", healthcheckAddress, err)
		return false
	}
	tcpConn, err := dialTCP("tcp4", ip, port, timeout, fwmark)
	if err != nil {
		logging.WithFields(logrus.Fields{
			"entity":   tcpHealthcheckName,
			"event id": id,
		}).Tracef("tcp connect to %v error: %v", healthcheckAddress, err)
		return false
	}
	conn := net.Conn(tcpConn)
	defer conn.Close()

	if conn != nil {
		logging.WithFields(logrus.Fields{
			"entity":   tcpHealthcheckName,
			"event id": id,
		}).Tracef("tcp connect to %v ok", healthcheckAddress)
		return true
	}

	// somehow it can be..
	logging.WithFields(logrus.Fields{
		"entity":   tcpHealthcheckName,
		"event id": id,
	}).Error("tcp connect unknown error: connection is nil, but have no errors")
	return false
}
