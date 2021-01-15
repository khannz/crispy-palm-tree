package portadapter

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/sirupsen/logrus"
)

type HttpAdvancedEntity struct {
	logging *logrus.Logger
}

func NewHttpAdvancedEntity(logging *logrus.Logger) *HttpAdvancedEntity {
	return &HttpAdvancedEntity{logging: logging}
}

const httpAdvancedHealthcheckName = "http advanced healthcheck"

// UnknownDataStruct used for trying to get an unknown json or array of json's
type UnknownDataStruct struct {
	UnknowMap         map[string]string
	UnknowArrayOfMaps []map[string]string
}

func (httpAdvancedEntity *HttpAdvancedEntity) IsHttpAdvancedCheckOk(healthcheckType string,
	healthcheckAddress string,
	nearFieldsMode bool,
	userDefinedData map[string]string,
	timeout time.Duration,
	fwmark int,
	id string) bool {
	switch healthcheckType {
	case "http-advanced-json":
		return httpAdvancedJSONCheckOk(healthcheckAddress, nearFieldsMode,
			userDefinedData, timeout, fwmark, id, httpAdvancedEntity.logging)
	default:
		httpAdvancedEntity.logging.WithFields(logrus.Fields{
			"entity":   httpAdvancedHealthcheckName,
			"event id": id,
		}).Errorf("Heathcheck error: http advanced check fail error: unknown check type: %v", healthcheckType)
		return false
	}
}

func httpAdvancedJSONCheckOk(healthcheckAddress string,
	nearFieldsMode bool,
	userDefinedData map[string]string,
	timeout time.Duration,
	fwmark int,
	id string,
	logging *logrus.Logger) bool {
	method := "GET"
	request := "/"
	responseCode := 200
	u, err := url.Parse(request)
	if err != nil {
		return false
	}
	u.Scheme = "https"

	// FIXME: need to fix hc models
	u.Host = healthcheckAddress
	// if u.Host == "" {
	// 	u.Host = ipS + ":" + port
	// }
	ip, port, err := net.SplitHostPort(healthcheckAddress)
	if err != nil {
		logging.WithFields(logrus.Fields{
			"entity":   httpsHealthcheckName,
			"event id": id,
		}).Errorf("https can't SplitHostPort from %v: %v", healthcheckAddress, err)
		return false
	}

	// proxy := (func(*http.Request) (*url.URL, error))(nil)

	var dialer func(network, addr string) (net.Conn, error)
	network := "tcp4"
	// Both DSR and TUN mode requires socket marks
	tcpConn, err := dialTCP(network, ip, port, timeout, fwmark)
	if err != nil {
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
	req, err := http.NewRequest(method, request, nil)
	if err != nil {
		logging.WithFields(logrus.Fields{
			"entity":   httpsHealthcheckName,
			"event id": id,
		}).Errorf("http can't create request for %v: %v", healthcheckAddress, err)
		return false
	}
	req.URL = u
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		logging.WithFields(logrus.Fields{
			"entity":   httpAdvancedHealthcheckName,
			"event id": id,
		}).Tracef("http advanced JSON check for %v error: %v", healthcheckAddress, err)
		return false
	}

	// Check response code.
	var codeOk bool
	if responseCode == 0 {
		codeOk = true
	} else if resp.StatusCode == responseCode {
		codeOk = true
	} else {
		return false
	}

	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logging.WithFields(logrus.Fields{
			"entity":   httpAdvancedHealthcheckName,
			"event id": id,
		}).Tracef("Heathcheck error: Read http response errror: %v", err)
		return false
	}

	uds := UnknownDataStruct{}
	if err := json.Unmarshal(response, &uds.UnknowMap); err != nil {
		if err := json.Unmarshal(response, &uds.UnknowArrayOfMaps); err != nil {
			logging.WithFields(logrus.Fields{
				"entity":   httpAdvancedHealthcheckName,
				"event id": id,
			}).Tracef("Heathcheck error: http advanced JSON check fail error: can't unmarshal response from: %v, error: %v",
				healthcheckAddress,
				err)
			return false
		}
	}

	if uds.UnknowMap == nil && uds.UnknowArrayOfMaps == nil {
		logging.WithFields(logrus.Fields{
			"entity":   httpAdvancedHealthcheckName,
			"event id": id,
		}).Tracef("Heathcheck error: http advanced JSON check fail error: response is nil from: %v", healthcheckAddress)
		return false
	}

	if nearFieldsMode { // mode for finding all matches for the desired object in a single map
		if isFinderForNearFieldsModeFail(userDefinedData, uds, healthcheckAddress, id, logging) { // if false do not return, continue range params
			return false
		}
	} else {
		if isFinderMapToMapFail(userDefinedData, uds, healthcheckAddress, id, logging) { // if false do not return, continue range params
			return false
		}
	}

	return codeOk
}

func isFinderForNearFieldsModeFail(userSearchData map[string]string,
	unknownDataStruct UnknownDataStruct,
	healthcheckAddres string,
	id string,
	logging *logrus.Logger) bool {
	numberOfRequiredMatches := len(userSearchData) // the number of required matches in the user's search map
	var mapForSearch map[string]string             // the map that we will use to search for all matches(beacose that nearFieldsMode)
	for sK, sV := range userSearchData {           // go through the search map
		if numberOfRequiredMatches != 0 { // checking that not all matches are found within the search map
			if mapForSearch != nil {
				if isKVequal(sK, sV, mapForSearch) {
					numberOfRequiredMatches-- // reduced search by length of matches
				}
			} else { // If matches haven't been found yet (nearFieldsMode)
				if unknownDataStruct.UnknowArrayOfMaps != nil {
					for _, incomeData := range unknownDataStruct.UnknowArrayOfMaps { // go through the array of maps received on request
						if isKVequal(sK, sV, incomeData) {
							numberOfRequiredMatches-- // reduced search by length of matches
							mapForSearch = incomeData // other matches for the desired map will be searched only in this one (nearFieldsMode)
							break
						}
					}
				} else if unknownDataStruct.UnknowMap != nil {
					if isKVequal(sK, sV, unknownDataStruct.UnknowMap) {
						numberOfRequiredMatches--                  // reduced search by length of matches
						mapForSearch = unknownDataStruct.UnknowMap // other matches for the desired map will be searched only in this one (nearFieldsMode)
					}
				}
			}
		}
	}
	if numberOfRequiredMatches != 0 {
		logging.WithFields(logrus.Fields{
			"entity":   httpAdvancedHealthcheckName,
			"event id": id,
		}).Tracef("Heathcheck http advanded json for %v failed: not all required data finded", healthcheckAddres)
		return true
	}
	logging.WithFields(logrus.Fields{
		"entity":   httpAdvancedHealthcheckName,
		"event id": id,
	}).Tracef("Heathcheck http advanded json for %v succes", healthcheckAddres)

	return false
}

func isFinderMapToMapFail(userSearchData map[string]string,
	unknownDataStruct UnknownDataStruct,
	healthcheckAddres string,
	id string,
	logging *logrus.Logger) bool {
	numberOfRequiredMatches := len(userSearchData) // the number of required matches in the user's search map

	for sK, sV := range userSearchData { // go through the search map
		if numberOfRequiredMatches != 0 { // checking that not all matches are found within the search map
			if unknownDataStruct.UnknowArrayOfMaps != nil {
				for _, incomeData := range unknownDataStruct.UnknowArrayOfMaps { // go through the array of maps received on request
					if isKVequal(sK, sV, incomeData) {
						numberOfRequiredMatches-- // reduced search by length of matches
						break
					}
				}
			} else if unknownDataStruct.UnknowMap != nil {
				if isKVequal(sK, sV, unknownDataStruct.UnknowMap) {
					numberOfRequiredMatches-- // reduced search by length of matches
				}
			}
		}
	}

	if numberOfRequiredMatches != 0 {
		logging.WithFields(logrus.Fields{
			"entity":   httpAdvancedHealthcheckName,
			"event id": id,
		}).Tracef("Heathcheck http advanded json for %v failed: not all required data finded", healthcheckAddres)

		return true
	}
	logging.WithFields(logrus.Fields{
		"entity":   httpAdvancedHealthcheckName,
		"event id": id,
	}).Tracef("Heathcheck http advanded json for %v succes", healthcheckAddres)

	return false
}

func isKVequal(k string, v interface{}, mapForSearch map[string]string) bool {
	if mI, isKeyFinded := mapForSearch[k]; isKeyFinded {
		if v == mI {
			return true
		}
	}
	return false
}
