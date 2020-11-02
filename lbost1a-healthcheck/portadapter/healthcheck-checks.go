package portadapter

import (
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

func (hc *HeathcheckEntity) tcpCheckOk(healthcheckAddress string, timeout time.Duration, id string) bool {
	hcSlice := strings.Split(healthcheckAddress, ":")
	hcPort := ""
	if len(hcSlice) > 1 {
		hcPort = hcSlice[1]
	}
	hcIP := hcSlice[0]
	dialer := net.Dialer{
		LocalAddr: hc.techInterface,
		Timeout:   timeout}

	conn, err := dialer.Dial("tcp", net.JoinHostPort(hcIP, hcPort))
	if err != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":   healthcheckName,
			"event id": id,
		}).Tracef("Heathcheck error: Connecting tcp connect error: %v", err)
		return false
	}
	defer conn.Close()

	if conn != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":   healthcheckName,
			"event id": id,
		}).Tracef("Heathcheck info port opened: %v", net.JoinHostPort(hcIP, hcPort))
		return true
	}

	// somehow it can be..
	hc.logging.WithFields(logrus.Fields{
		"entity":   healthcheckName,
		"event id": id,
	}).Error("Heathcheck has unknown error: connection is nil, but have no errors")
	return false
}

func (hc *HeathcheckEntity) httpCheckOk(healthcheckAddress string, timeout time.Duration, id string) bool {
	// FIXME: https checks also here
	roundTripper := &http.Transport{
		Dial: (&net.Dialer{
			LocalAddr: hc.techInterface,
			Timeout:   timeout,
		}).Dial,
		TLSHandshakeTimeout: timeout * 6 / 10, // hardcode
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
	}
	client := http.Client{
		Transport: roundTripper,
		Timeout:   timeout,
	}
	resp, err := client.Get(healthcheckAddress)
	if err != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":   healthcheckName,
			"event id": id,
		}).Tracef("Heathcheck error: Connecting http error: %v", err)
		return false
	}

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":   healthcheckName,
			"event id": id,
		}).Tracef("Heathcheck error: Read http response errror: %v", err)
		return false
	}
	return true
}

func (hc *HeathcheckEntity) icmpCheckOk(healthcheckAddress string, timeout time.Duration, id string) bool {
	// Start listening for icmp replies
	icpmConnection, err := icmp.ListenPacket("ip4:icmp", hc.techInterface.String())
	if err != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":   healthcheckName,
			"event id": id,
		}).Tracef("Heathcheck error: icpm connection error: %v", err)
		return false
	}
	defer icpmConnection.Close()

	// Get the real IP of the target
	dst, err := net.ResolveIPAddr("ip4", healthcheckAddress)
	if err != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":   healthcheckName,
			"event id": id,
		}).Tracef("Heathcheck error: icpm resolve ip addr error: %v", err)
		return false
	}

	m := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   1,
			Seq:  1,
			Data: []byte("hello")},
	}

	b, err := m.Marshal(nil)
	if err != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":   healthcheckName,
			"event id": id,
		}).Tracef("Heathcheck error: icpm marshall message error: %v", err)
		return false
	}

	// Send it
	n, err := icpmConnection.WriteTo(b, dst)
	if err != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":   healthcheckName,
			"event id": id,
		}).Tracef("Heathcheck error: icpm write bytes to error: %v", err)
		return false
	} else if n != len(b) {
		hc.logging.WithFields(logrus.Fields{
			"entity":   healthcheckName,
			"event id": id,
		}).Tracef("Heathcheck error: icpm write bytes to error (not all of bytes was send): %v", err)
		return false
	}

	// Wait for a reply
	reply := make([]byte, 1500)
	err = icpmConnection.SetReadDeadline(time.Now().Add(timeout))
	if err != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":   healthcheckName,
			"event id": id,
		}).Tracef("Heathcheck error: icpm set read deadline error: %v", err)
		return false
	}
	n, peer, err := icpmConnection.ReadFrom(reply)
	if err != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":   healthcheckName,
			"event id": id,
		}).Tracef("Heathcheck error: icpm read reply error: %v", err)
		return false
	}

	// Let's look what we have in reply
	rm, err := icmp.ParseMessage(protocolICMP, reply[:n])
	if err != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":   healthcheckName,
			"event id": id,
		}).Tracef("Heathcheck error: icpm parse message error: %v", err)
		return false
	}
	switch rm.Type {
	case ipv4.ICMPTypeEchoReply:
		hc.logging.WithFields(logrus.Fields{
			"entity":   healthcheckName,
			"event id": id,
		}).Tracef("Heathcheck icpm for %v succes", healthcheckAddress)
		return true
	default:
		hc.logging.WithFields(logrus.Fields{
			"entity":   healthcheckName,
			"event id": id,
		}).Tracef("Heathcheck error: icpm for %v reply type error: got %+v from %v; want echo reply",
			healthcheckAddress,
			rm,
			peer)
		return false
	}
}

// http advanced start
func (hc *HeathcheckEntity) httpAdvancedCheckOk(hcType string,
	hcAddress string,
	nearFieldsMode bool,
	userDefinedData map[string]string,
	timeout time.Duration,
	id string) bool {
	switch hcType {
	case "http-advanced-json":
		return hc.httpAdvancedJSONCheckOk(hcAddress, nearFieldsMode,
			userDefinedData, timeout, id)
	default:
		hc.logging.WithFields(logrus.Fields{
			"entity":   healthcheckName,
			"event id": id,
		}).Errorf("Heathcheck error: http advanced check fail error: unknown check type: %v", hcType)
		return false
	}
}

func (hc *HeathcheckEntity) httpAdvancedJSONCheckOk(hcAddress string,
	nearFieldsMode bool,
	userDefinedData map[string]string,
	timeout time.Duration,
	id string) bool {
	client := http.Client{
		Timeout: timeout,
	}

	req, err := http.NewRequest("GET", hcAddress, nil)
	if err != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":   healthcheckName,
			"event id": id,
		}).Errorf("Heathcheck error: http advanced JSON check fail error: can't make new http request: %v", err)
		return false
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":   healthcheckName,
			"event id": id,
		}).Tracef("Heathcheck error: Connecting http advanced JSON check error: %v", err)
		return false
	}

	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":   healthcheckName,
			"event id": id,
		}).Tracef("Heathcheck error: Read http response errror: %v", err)
		return false
	}

	u := UnknownDataStruct{}
	if err := json.Unmarshal(response, &u.UnknowMap); err != nil {
		if err := json.Unmarshal(response, &u.UnknowArrayOfMaps); err != nil {
			hc.logging.WithFields(logrus.Fields{
				"entity":   healthcheckName,
				"event id": id,
			}).Tracef("Heathcheck error: http advanced JSON check fail error: can't unmarshal response from: %v, error: %v",
				hcAddress,
				err)
			return false
		}
	}

	if u.UnknowMap == nil && u.UnknowArrayOfMaps == nil {
		hc.logging.WithFields(logrus.Fields{
			"entity":   healthcheckName,
			"event id": id,
		}).Tracef("Heathcheck error: http advanced JSON check fail error: response is nil from: %v", hcAddress)
		return false
	}

	if nearFieldsMode { // mode for finding all matches for the desired object in a single map
		if hc.isFinderForNearFieldsModeFail(userDefinedData, u, hcAddress, id) { // if false do not return, continue range params
			return false
		}
	} else {
		if hc.isFinderMapToMapFail(userDefinedData, u, hcAddress, id) { // if false do not return, continue range params
			return false
		}
	}

	return true
}

func (hc *HeathcheckEntity) isFinderForNearFieldsModeFail(userSearchData map[string]string,
	unknownDataStruct UnknownDataStruct,
	healthcheckAddres string,
	id string) bool {
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
		hc.logging.WithFields(logrus.Fields{
			"entity":   healthcheckName,
			"event id": id,
		}).Tracef("Heathcheck http advanded json for %v failed: not all required data finded", healthcheckAddres)
		return true
	}
	hc.logging.WithFields(logrus.Fields{
		"entity":   healthcheckName,
		"event id": id,
	}).Tracef("Heathcheck http advanded json for %v succes", healthcheckAddres)

	return false
}

func (hc *HeathcheckEntity) isFinderMapToMapFail(userSearchData map[string]string,
	unknownDataStruct UnknownDataStruct,
	healthcheckAddres string,
	id string) bool {
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
		hc.logging.WithFields(logrus.Fields{
			"entity":   healthcheckName,
			"event id": id,
		}).Tracef("Heathcheck http advanded json for %v failed: not all required data finded", healthcheckAddres)

		return true
	}
	hc.logging.WithFields(logrus.Fields{
		"entity":   healthcheckName,
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

// http advanced end
