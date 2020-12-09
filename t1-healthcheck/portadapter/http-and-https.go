package portadapter

import (
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

const httpsHealthcheckName = "https healthcheck"

type HttpsAndHttpsEntity struct {
	logging *logrus.Logger
}

func NewHttpsAndHttpsEntity(logging *logrus.Logger) *HttpsAndHttpsEntity {
	return &HttpsAndHttpsEntity{logging: logging}
}

func (httpsAndhttpsEntity *HttpsAndHttpsEntity) IsHttpOrHttpsCheckOk(healthcheckAddress string,
	timeout time.Duration,
	fwmark int,
	isHttpCheck bool,
	id string) bool {
	// method := "GET" // TODO: remove hardcode
	request := "/index.html"
	responseCodes := make(map[int]struct{})
	responseCodes[200] = struct{}{}
	responseCodes[301] = struct{}{}
	responseCodes[307] = struct{}{}
	responseCodes[308] = struct{}{}

	var newHCAddress string
	if isHttpCheck {
		newHCAddress = "http://" + healthcheckAddress + request
	} else {
		newHCAddress = "https://" + healthcheckAddress + request
	}

	ip, port, err := net.SplitHostPort(healthcheckAddress)
	if err != nil {
		httpsAndhttpsEntity.logging.WithFields(logrus.Fields{
			"entity":   httpsHealthcheckName,
			"event id": id,
		}).Errorf("https can't SplitHostPort from %v: %v", healthcheckAddress, err)
		return false
	}

	var dialer func(network, addr string) (net.Conn, error)
	network := "tcp4"
	// Both DSR and TUN mode requires socket marks
	tcpConn, err := dialTCP(network, ip, port, timeout, fwmark)
	if err != nil {
		httpsAndhttpsEntity.logging.WithFields(logrus.Fields{
			"entity":   httpsHealthcheckName,
			"event id": id,
		}).Debugf("can't connect from %v: %v", healthcheckAddress, err)
		return false
	}
	defer tcpConn.f.Close()
	defer tcpConn.netConn.Close()

	dialer = func(net string, addr string) (net.Conn, error) {
		return tcpConn.netConn, nil
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return errors.New("redirect not permitted")
		},
		Transport: &http.Transport{
			Dial: dialer,
			// Proxy:           proxy,
			TLSClientConfig: tlsConfig,
		},
		Timeout: timeout,
	}

	// If we received a response we want to process it, even in the
	// presence of an error - a redirect 3xx will result in both the
	// response and an error being returned.
	resp, err := client.Get(newHCAddress)
	if err != nil {
		httpsAndhttpsEntity.logging.WithFields(logrus.Fields{
			"entity":   httpsHealthcheckName,
			"event id": id,
		}).Tracef("Heathcheck error: Connecting http(https) error: %v", err)
		if resp != nil {
			if resp.Body != nil {
				defer resp.Body.Close()
			}
			_, codeOk := responseCodes[resp.StatusCode]
			return codeOk
		}
		return false
	}

	if resp == nil {
		return false
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}

	// Check response code.
	// var codeOk bool
	// switch responseCodes // TODO:
	_, codeOk := responseCodes[resp.StatusCode]
	// Check response body.
	bodyOk := true

	return codeOk && bodyOk
}
