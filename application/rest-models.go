package application

import "time"

// NewServiceInfo ...
type NewServiceInfo struct {
	ID                 string              `json:"id" validate:"uuid4" example:"7a7aebea-4e05-45b9-8d11-c4115dbdd4a2"`
	ServiceIP          string              `json:"serviceIP" validate:"ipv4" example:"1.1.1.1"`
	ServicePort        string              `json:"servicePort" validate:"required" example:"1111"`
	Healtcheck         ServiceHealthcheck  `json:"Healtcheck" validate:"required"`
	ApplicationServers []ServerApplication `json:"applicationServers" validate:"required,dive,required"`
	BalanceType        string              `json:"balanceType" validate:"required" example:"rr"`
	RoutingType        string              `json:"routingType" validate:"required" example:"masquarading,tunneling"`
}

// GetServiceStateRequest ...
type GetServiceStateRequest struct {
	ID          string `json:"id" validate:"uuid4" example:"7a7aebea-4e05-45b9-8d11-c4115dbdd4a2"`
	ServiceIP   string `json:"serviceIP" validate:"ipv4" example:"1.1.1.1"`
	ServicePort string `json:"servicePort" validate:"required" example:"1111"`
}

// GetAllServicesRequest ...
type GetAllServicesRequest struct {
	ID string `json:"id" validate:"uuid4" example:"7a7aebea-4e05-45b9-8d11-c4115dbdd4a2"`
}

// GetAllServicesResponse ...
type GetAllServicesResponse struct {
	ID                       string                        `json:"id"`
	JobCompletedSuccessfully bool                          `json:"jobCompletedSuccessfully"`
	AllServices              []UniversalResponseWithStates `json:"allServices,omitempty"`
	ExtraInfo                string                        `json:"extraInfo,omitempty"`
}

// ModifyServiceInfo ...
type ModifyServiceInfo struct {
	ID                 string              `json:"id" validate:"uuid4" example:"7a7aebea-4e05-45b9-8d11-c4115dbdd4a2"`
	ServiceIP          string              `json:"serviceIP" validate:"ipv4" example:"1.1.1.1"`
	ServicePort        string              `json:"servicePort" validate:"required" example:"1111"`
	Healtcheck         ServiceHealthcheck  `json:"Healtcheck" validate:"required"`
	ApplicationServers []ServerApplication `json:"applicationServers" validate:"required,dive,required"`
	BalanceType        string              `json:"balanceType" validate:"required" example:"rr"`
	RoutingType        string              `json:"routingType" validate:"required" example:"masquarading,tunneling"`
}

// RemoveApplicationServersRequest ...
type RemoveApplicationServersRequest struct {
	ID                 string              `json:"id" validate:"uuid4" example:"7a7aebea-4e05-45b9-8d11-c4115dbdd4a2"`
	ServiceIP          string              `json:"serviceIP" validate:"ipv4" example:"1.1.1.1"`
	ServicePort        string              `json:"servicePort" validate:"required" example:"1111"`
	ApplicationServers []ServerApplication `json:"applicationServers" validate:"required,dive,required"`
}

// RemoveServiceInfo ...
type RemoveServiceInfo struct {
	ID          string `json:"id" validate:"uuid4" example:"7a7aebea-4e05-45b9-8d11-c4115dbdd4a2"`
	ServiceIP   string `json:"serviceIP" validate:"ipv4" example:"1.1.1.1"`
	ServicePort string `json:"servicePort" validate:"required" example:"1111"`
}

// ServiceHealthcheck ...
type ServiceHealthcheck struct {
	Type                 string        `json:"type" validate:"required" example:"tcp"`
	Timeout              time.Duration `json:"timeout" validate:"required" example:"1000000000"`
	RepeatHealthcheck    time.Duration `json:"repeatHealthcheck" validate:"required" example:"3000000000"`
	PercentOfAlivedForUp int           `json:"percentOfAlivedForUp" validate:"required,gt=0,lte=100"`
}

// AdvancedHealthcheckParameters ...
type AdvancedHealthcheckParameters struct {
	NearFieldsMode  bool                   `json:"nearFieldsMode" validate:"required" example:"false"`
	UserDefinedData map[string]interface{} `json:"userDefinedData" validate:"required"`
}

// ServerHealthcheck ...
type ServerHealthcheck struct {
	TypeOfCheck                   string                          `json:"typeOfCheck,omitempty" example:"http-advanced-json"`
	HealthcheckAddress            string                          `json:"healthcheckAddress,omitempty"` // FIXME: need extra validate; ip+port, http address or some one else
	AdvancedHealthcheckParameters []AdvancedHealthcheckParameters `json:"advancedHealthcheckParameters"`
}

// ServerApplication ...
type ServerApplication struct {
	ServerIP                    string            `json:"ip" validate:"required,ipv4" example:"1.1.1.1"`
	ServerPort                  string            `json:"port" validate:"required" example:"1111"`
	ServerHealthcheck           ServerHealthcheck `json:"serverHealthcheck,omitempty"`
	ServerСonfigurationCommands string            `json:"bashCommands,omitempty" swaggerignore:"true"` // FIXME: ingore gin!
}

// UniversalResponse ...
type UniversalResponse struct {
	ID                       string              `json:"id,omitempty"`
	ApplicationServers       []ServerApplication `json:"applicationServers,omitempty"`
	ServiceIP                string              `json:"serviceIP,omitempty"`
	ServicePort              string              `json:"servicePort,omitempty"`
	Healthcheck              ServiceHealthcheck  `json:"healthcheck,omitempty"`
	JobCompletedSuccessfully bool                `json:"jobCompletedSuccessfully"`
	ExtraInfo                string              `json:"extraInfo,omitempty"`
	BalanceType              string              `json:"balanceType,omitempty"`
	RoutingType              string              `json:"routingType,omitempty"`
}

// ServerApplicationWithStates ...
type ServerApplicationWithStates struct {
	ServerIP                    string            `json:"ip" validate:"required,ipv4" example:"1.1.1.1"`
	ServerPort                  string            `json:"port" validate:"required" example:"1111"`
	ServerHealthcheck           ServerHealthcheck `json:"serverHealthcheck,omitempty"`
	IsUp                        bool              `json:"serverIsUp"`
	ServerСonfigurationCommands string            `json:"bashCommands,omitempty" swaggerignore:"true"` // FIXME: ingore gin!
}

// UniversalResponseWithStates ...
type UniversalResponseWithStates struct {
	ID                       string                        `json:"id,omitempty"`
	ApplicationServers       []ServerApplicationWithStates `json:"applicationServers,omitempty"`
	ServiceIP                string                        `json:"serviceIP,omitempty"`
	ServicePort              string                        `json:"servicePort,omitempty"`
	Healthcheck              ServiceHealthcheck            `json:"healthcheck,omitempty"`
	JobCompletedSuccessfully bool                          `json:"jobCompletedSuccessfully"`
	ExtraInfo                string                        `json:"extraInfo,omitempty"`
	BalanceType              string                        `json:"balanceType,omitempty"`
	RoutingType              string                        `json:"routingType,omitempty"`
	IsUp                     bool                          `json:"serviceIsUp"`
}

// AddApplicationServersRequest ...
type AddApplicationServersRequest struct {
	ID                 string              `json:"id" validate:"uuid4" example:"7a7aebea-4e05-45b9-8d11-c4115dbdd4a2"`
	ServiceIP          string              `json:"serviceIP" validate:"ipv4" example:"1.1.1.1"`
	ServicePort        string              `json:"servicePort" validate:"required" example:"1111"`
	ApplicationServers []ServerApplication `json:"applicationServers" validate:"required,dive,required"`
}

// TokenRequest ...
type TokenRequest struct {
	User     string `json:"user" validate:"required" example:"Sneshana-IE"`
	Password string `json:"password" validate:"required" example:"secret-password"`
	ID       string `json:"id" validate:"required,uuid4" example:"7a7aebea-4e05-45b9-8d11-c4115dbdd4a2"`
}

// TokenResponseOkay ...
type TokenResponseOkay struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ID           string `json:"id"`
}

// TokenResponseError ...
type TokenResponseError struct {
	Error string `json:"error"`
	ID    string `json:"id"`
}
