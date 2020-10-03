package application

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

const addApplicationServersRequestName = "add application servers"

// addApplicationServers godoc
// @tags load balancer
// @Summary Add application servers
// @Description Beyond the network balance
// @Param addr path string true "IP"
// @Param port path uint true "Port"
// @Param incomeJSON body application.AddApplicationServersRequest true "Expected json"
// @Accept json
// @Produce json
// @Success 200 {object} application.Service "If all okay"
// @Failure 400 {string} error "Bad request"
// @Failure 500 {string} error "Internal error"
// @Router /service/{addr}/{port}/add-application-servers [post]
// // // @Security ApiKeyAuth
func (restAPI *RestAPIstruct) addApplicationServers(ginContext *gin.Context) {
	addApplicationServersRequestID := restAPI.balancerFacade.IDgenerator.NewID()
	restAPI.balancerFacade.Logging.WithFields(logrus.Fields{"event id": addApplicationServersRequestID}).Infof("got new %v request", addApplicationServersRequestName)
	addApplicationServersRequest := &AddApplicationServersRequest{}

	if err := ginContext.ShouldBindJSON(addApplicationServersRequest); err != nil {
		restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
			"event id": addApplicationServersRequestID,
		}).Errorf("can't %v, got error: %v", addApplicationServersRequestName, err)
		ginContext.String(http.StatusInternalServerError, "got internal error: %b"+err.Error())
		return
	}
	addApplicationServersRequest.IP = ginContext.Param("addr")
	addApplicationServersRequest.Port = ginContext.Param("port")

	if validateError := addApplicationServersRequest.validateAddApplicationServersRequest(); validateError != nil {
		restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
			"entity":   addApplicationServersRequestName,
			"event id": addApplicationServersRequestID,
		}).Errorf("invalide income request: %v", validateError.Error())

		ginContext.String(http.StatusBadRequest, "invalide income request: "+validateError.Error())
		return
	}

	preparedAddApplicationServersRequest := &Service{
		IP:                 addApplicationServersRequest.IP,
		Port:               addApplicationServersRequest.Port,
		ApplicationServers: addApplicationServersRequest.ApplicationServers,
	}

	updatedServiceInfo, err := restAPI.balancerFacade.AddApplicationServers(preparedAddApplicationServersRequest,
		addApplicationServersRequestID)
	if err != nil {
		restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
			"event id": addApplicationServersRequestID,
		}).Errorf("can't %v, got error: %v", addApplicationServersRequestName, err.Error())

		ginContext.String(http.StatusInternalServerError, err.Error())
		return
	}

	restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
		"event id": addApplicationServersRequestID,
	}).Infof("request %v done", addApplicationServersRequestName)

	ginContext.JSON(http.StatusOK, updatedServiceInfo)
}

func (addApplicationServersRequest *AddApplicationServersRequest) validateAddApplicationServersRequest() error {
	validate := validator.New()
	validate.RegisterStructValidation(customPortAddApplicationServersRequestValidation, AddApplicationServersRequest{})
	// validate.RegisterStructValidation(customPortServerApplicationValidation, ServerApplication{})
	// validate.RegisterStructValidation(customServiceHealthcheckValidation, ServiceHealthcheck{})
	if err := validate.Struct(addApplicationServersRequest); err != nil {
		return modifyValidateError(err)
	}
	return nil
}

func customPortAddApplicationServersRequestValidation(sl validator.StructLevel) {
	nbi := sl.Current().Interface().(AddApplicationServersRequest)
	port, err := strconv.Atoi(nbi.Port)
	if err != nil {
		sl.ReportError(nbi.Port, "servicePort", "Port", "port must be number", "")
	}
	if !(port > 0) || !(port < 20000) {
		sl.ReportError(nbi.Port, "servicePort", "Port", "port must gt=0 and lt=20000", "")
	}
}
