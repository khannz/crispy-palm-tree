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

type HttpsEntity struct {
	logging *logrus.Logger
}

func NewHttpsEntity(logging *logrus.Logger) *HttpsEntity {
	return &HttpsEntity{logging: logging}
}

func (httpsEntity *HttpsEntity) IsHttpsCheckOk(healthcheckAddress string,
	timeout time.Duration,
	fwmark int,
	id string) bool {
	method := "GET"
	request := "/"
	responseCode := 200
	u, err := url.Parse(request)
	if err != nil {
		return false
	}
	u.Scheme = "https"

	ip, port, err := net.SplitHostPort(healthcheckAddress)
	if err != nil {
		httpsEntity.logging.WithFields(logrus.Fields{
			"entity":   httpsHealthcheckName,
			"event id": id,
		}).Errorf("https can't SplitHostPort from %v: %v", healthcheckAddress, err)
		return false
	}

	// FIXME: need to fix hc models
	u.Host = ip
	// if u.Host == "" {
	// 	u.Host = ipS + ":" + port
	// }

	// proxy := (func(*http.Request) (*url.URL, error))(nil)

	var dialer func(network, addr string) (net.Conn, error)
	network := "tcp4"
	// Both DSR and TUN mode requires socket marks
	conn, err := dialTCP(network, ip, port, timeout, fwmark)
	if err != nil {
		httpsEntity.logging.WithFields(logrus.Fields{
			"entity":   httpsHealthcheckName,
			"event id": id,
		}).Debugf("can't connect from %v: %v", healthcheckAddress, err)
		return false
	}
	defer conn.Close()

	dialer = func(net string, addr string) (net.Conn, error) {
		return conn, nil
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
	req, err := http.NewRequest(method, request, nil)
	if err != nil {
		httpsEntity.logging.WithFields(logrus.Fields{
			"entity":   httpsHealthcheckName,
			"event id": id,
		}).Errorf("https can't create request for %v: %v", healthcheckAddress, err)
		return false
	}
	req.URL = u

	// If we received a response we want to process it, even in the
	// presence of an error - a redirect 3xx will result in both the
	// response and an error being returned.
	resp, err := client.Do(req)
	if err != nil {
		httpsEntity.logging.WithFields(logrus.Fields{
			"entity":   httpsHealthcheckName,
			"event id": id,
		}).Errorf("https response error: %v", err)
	}
	if resp == nil {
		return false
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}

	// Check response code.
	var codeOk bool
	if responseCode == 0 {
		codeOk = true
	} else if resp.StatusCode == responseCode {
		codeOk = true
	}

	// Check response body.
	bodyOk := true

	return codeOk && bodyOk
}
