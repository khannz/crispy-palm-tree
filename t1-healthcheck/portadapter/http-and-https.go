package portadapter

import (
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"net/url"
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
	uri string,
	validResponseCodes map[int]struct{},
	timeout time.Duration,
	fwmark int,
	isHttpCheck bool,
	id string) bool {
	// method := "GET" // TODO: remove hardcode

	var newHCAddress string
	if isHttpCheck {
		newHCAddress = "http://" + healthcheckAddress + uri
	} else {
		newHCAddress = "https://" + healthcheckAddress + uri
	}

	_, err := url.ParseRequestURI(newHCAddress)
	if err != nil {
		httpsAndhttpsEntity.logging.WithFields(logrus.Fields{
			"entity":   httpsHealthcheckName,
			"event id": id,
		}).Errorf("invalid url address. uri: %v, healthcheck address %v, full address: %v. error: %v",
			uri,
			healthcheckAddress,
			err)
		return false
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
	defer tcpConn.Close()

	dialer = func(net string, addr string) (net.Conn, error) {
		return tcpConn, nil
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
		}).Tracef("Healthcheck error: Connecting http(https) error: %v", err)
		if resp != nil {
			if resp.Body != nil {
				defer resp.Body.Close()
			}
			_, isCodeOk := validResponseCodes[resp.StatusCode]
			return isCodeOk
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
	_, codeOk := validResponseCodes[resp.StatusCode]
	// Check response body.
	bodyOk := true

	return codeOk && bodyOk
}
