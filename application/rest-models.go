package application

import "time"

type Service struct {
	IP                    string                 `json:"ip" validate:"ipv4" example:"1.1.1.1"`
	Port                  string                 `json:"port" validate:"required" example:"1111"`
	IsUp                  bool                   `json:"isUp,omitempty"`
	BalanceType           string                 `json:"balanceType" validate:"required" example:"rr"`
	RoutingType           string                 `json:"routingType" validate:"required" example:"masquerading,tunneling"`
	Protocol              string                 `json:"protocol" validate:"required" example:"tcp,udp"`
	AlivedAppServersForUp int                    `json:"alivedAppServersForUp" validate:"required,gt=0,lte=100"`
	HCType                string                 `json:"hcType" validate:"required" example:"tcp"`
	HCRepeat              time.Duration          `json:"hcRepeat" validate:"required" example:"3000000000"`
	HCTimeout             time.Duration          `json:"hcTimeout" validate:"required" example:"1000000000"`
	HCNearFieldsMode      bool                   `json:"hcNearFieldsMode,omitempty"`
	HCUserDefinedData     map[string]interface{} `json:"hcUserDefinedData,omitempty"`
	HCRetriesForUP        int                    `json:"hcRetriesForUP" validate:"required,gt=0" example:"3"`
	HCRetriesForDown      int                    `json:"hcRetriesForDown" validate:"required,gt=0" example:"10"`
	ApplicationServers    []*ApplicationServer   `json:"applicationServers" validate:"required,dive,required"`
}

type ApplicationServer struct {
	IP                  string `json:"ip" validate:"ipv4" example:"1.1.1.1"`
	Port                string `json:"port" validate:"required" example:"1111"`
	IsUp                bool   `json:"isUp,omitempty"`
	HCAddress           string `json:"hcAddress" validate:"required" example:"http://1.1.1.1:1234"`
	ExampleBashCommands string `json:"exampleBashCommands,omitempty" swagger:"ignoreParam"`
}

// RemoveApplicationServersRequest ...
type RemoveApplicationServersRequest struct {
	IP                 string               `json:"serviceIP" validate:"ipv4" example:"1.1.1.1"`
	Port               string               `json:"servicePort" validate:"required" example:"1111"`
	ApplicationServers []*ApplicationServer `json:"applicationServers" validate:"required,dive,required"`
}

// AddApplicationServersRequest ...
type AddApplicationServersRequest struct {
	IP                 string               `json:"serviceIP" validate:"ipv4" example:"1.1.1.1"`
	Port               string               `json:"servicePort" validate:"required" example:"1111"`
	ApplicationServers []*ApplicationServer `json:"applicationServers" validate:"required,dive,required"`
}

// LoginRequest ...
type LoginRequest struct {
	User     string `json:"user" validate:"required" example:"Sneshana-IE"`
	Password string `json:"password" validate:"required" example:"secret-password"`
}

// LoginResponseOkay ...
type LoginResponseOkay struct {
	AccessToken string `json:"accessToken"`
}

// LoginResponseError ...
type LoginResponseError struct {
	Error string `json:"error"`
}
