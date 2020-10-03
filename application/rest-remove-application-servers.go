package application

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

const removeApplicationServersRequestName = "remove application servers"

// removeApplicationServers godoc
// @tags load balancer
// @Summary Remove application servers
// @Description Beyond the network balance
// @Param addr path string true "IP"
// @Param port path uint true "Port"
// @Param incomeJSON body application.RemoveApplicationServersRequest true "Expected json"
// @Accept json
// @Produce json
// @Success 200 {object} application.Service "If all okay"
// @Failure 400 {string} error "Bad request"
// @Failure 500 {string} error "Internal error"
// @Router /service/{addr}/{port}/remove-application-servers [post]
// // @Security ApiKeyAuth
func (restAPI *RestAPIstruct) removeApplicationServers(ginContext *gin.Context) {
	removeApplicationServersRequestID := restAPI.balancerFacade.IDgenerator.NewID()

	restAPI.balancerFacade.Logging.WithFields(logrus.Fields{"event id": removeApplicationServersRequestID}).Infof("got new %v request", removeApplicationServersRequestName)

	removeApplicationServersRequest := &RemoveApplicationServersRequest{}

	if err := ginContext.ShouldBindJSON(removeApplicationServersRequest); err != nil {
		restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
			"event id": removeApplicationServersRequestID,
		}).Errorf("can't %v, got error: %v", removeApplicationServersRequestName, err)
		ginContext.String(http.StatusInternalServerError, "got internal error: %b"+err.Error())
		return
	}
	removeApplicationServersRequest.IP = ginContext.Param("addr")
	removeApplicationServersRequest.Port = ginContext.Param("port")

	if validateError := removeApplicationServersRequest.validateRemoveApplicationServersRequest(); validateError != nil {
		restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
			"entity":   removeApplicationServersRequestName,
			"event id": removeApplicationServersRequestID,
		}).Errorf("validate fail for income nwb request: %v", validateError.Error())

		ginContext.String(http.StatusBadRequest, validateError.Error())
		return
	}

	preparedRemoveApplicationServersRequest := &Service{
		IP:                 removeApplicationServersRequest.IP,
		Port:               removeApplicationServersRequest.Port,
		ApplicationServers: removeApplicationServersRequest.ApplicationServers,
	}

	updatedServiceInfo, err := restAPI.balancerFacade.RemoveApplicationServers(preparedRemoveApplicationServersRequest,
		removeApplicationServersRequestID)
	if err != nil {
		restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
			"event id": removeApplicationServersRequestID,
		}).Errorf("can't %v, got error: %v", removeApplicationServersRequestName, err.Error())

		ginContext.String(http.StatusInternalServerError, err.Error())
		return
	}

	restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
		"event id": removeApplicationServersRequestID,
	}).Infof("request %v done", removeApplicationServersRequestName)

	ginContext.JSON(http.StatusOK, updatedServiceInfo)
}

func (removeApplicationServersRequest *RemoveApplicationServersRequest) validateRemoveApplicationServersRequest() error {
	validate := validator.New()
	validate.RegisterStructValidation(customPortRemoveApplicationServersRequestValidation, RemoveApplicationServersRequest{})
	// validate.RegisterStructValidation(customPortServerApplicationValidation, ServerApplication{})
	if err := validate.Struct(removeApplicationServersRequest); err != nil {
		return modifyValidateError(err)
	}
	return nil
}

func customPortRemoveApplicationServersRequestValidation(sl validator.StructLevel) {
	nbi := sl.Current().Interface().(RemoveApplicationServersRequest)
	port, err := strconv.Atoi(nbi.Port)
	if err != nil {
		sl.ReportError(nbi.Port, "servicePort", "Port", "port must be number", "")
	}
	if !(port > 0) || !(port < 20000) {
		sl.ReportError(nbi.Port, "servicePort", "Port", "port must gt=0 and lt=20000", "")
	}
}
