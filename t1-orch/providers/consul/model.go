package consul

import (
	consulAPI "github.com/hashicorp/consul/api"
	"github.com/khannz/crispy-palm-tree/t1-orch/providers"
)

type (
	//ServiceTransport main entity from Consul
	ServiceTransport struct {
		IP              string `json:"ip"`
		Port            string `json:"port"`
		BalanceType     string `json:"balanceType"`
		RoutingType     string `json:"routingType"`
		Protocol        string `json:"protocol"`
		HealthcheckType string `json:"healthcheckType"`
		HelloTimer      string `json:"helloTimer"`
		ResponseTimer   string `json:"responseTimer"`
		AliveThreshold  string `json:"aliveThreshold"`
		DeadThreshold   string `json:"deadThreshold"`
		Quorum          string `json:"quorum"`
		// Hysteresis      string `json:"hysteresis"`
		ApplicationServersTransport []ApplicationServerTransport `json:"-"`
		Uri                         string                       `json:"uri"`                // only for http(s) hc types
		ValidResponseCodes          []int64                      `json:"validResponseCodes"` // only for http(s) hc types
	}

	//ApplicationServerTransport ...
	ApplicationServerTransport struct {
		IP                 string `json:"ip"`
		Port               string `json:"port"`
		HealthcheckAddress string `json:"healthcheckAddress"`
	}

	//ServiceTransportData ...
	ServiceTransportData struct {
		providers.ServicesConfig `json:"-"`
		QueryMeta                *consulAPI.QueryMeta `json:"-"`
		Payload                  payload              `json:"-"`
	}

	payload interface {
		isPayloadPrivate()
	}

	//ServicesPayload services transport payload
	ServicesPayload struct {
		payload
		Services []ServiceTransport `json:"-"`
	}
	//NonePayload no payload
	NonePayload struct {
		payload
	}
)
