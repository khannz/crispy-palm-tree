package application

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

const newServiceRequestName = "new service"

// newService godoc
// @tags load balancer
// @Summary New service
// @Description Beyond the network balance
// @Param addr path string true "IP"
// @Param port path uint true "Port"
// @Param incomeJSON body application.Service true "Expected json"
// @Accept json
// @Produce json
// @Success 200 {object} application.Service "If all okay"
// @Failure 400 {string} error "Bad request"
// @Failure 500 {string} error "Internal error"
// @Router /service/{addr}/{port} [post]
// // // @Security ApiKeyAuth
func (restAPI *RestAPIstruct) newService(ginContext *gin.Context) {
	newServiceID := restAPI.balancerFacade.IDgenerator.NewID()
	restAPI.balancerFacade.Logging.WithFields(logrus.Fields{"event id": newServiceID}).Infof("got new %v request", newServiceRequestName)
	newService := &Service{}

	if err := ginContext.ShouldBindJSON(newService); err != nil {
		restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
			"event id": newServiceID,
		}).Errorf("can't %v, got error: %v", newServiceRequestName, err)
		ginContext.String(http.StatusInternalServerError, "got internal error: %b"+err.Error())
		return
	}
	newService.IP = ginContext.Param("addr")
	newService.Port = ginContext.Param("port")

	if validateError := newService.validateNewService(); validateError != nil {
		restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
			"entity":   newServiceRequestName,
			"event id": newServiceID,
		}).Errorf("validate fail for income nwb request: %v", validateError.Error())

		ginContext.String(http.StatusBadRequest, validateError.Error())
		return
	}

	nwbServiceInfo, err := restAPI.balancerFacade.NewService(newService,
		newServiceID)
	if err != nil {
		restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
			"event id": newServiceID,
		}).Errorf("can't %v, got error: %v", newServiceRequestName, err)

		ginContext.String(http.StatusInternalServerError, "got internal error: %b"+err.Error())
		return
	}

	restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
		"event id": newServiceID,
	}).Infof("request %v done", newServiceRequestName)

	ginContext.JSON(http.StatusOK, nwbServiceInfo)
}

func (newService *Service) validateNewService() error {
	validate := validator.New()
	// FIXME: return validate logic
	// validate.RegisterStructValidation(customPortValidationForCreateService, NewServiceInfo{})
	// validate.RegisterStructValidation(customPortServerApplicationValidation, ServerApplication{})
	// validate.RegisterStructValidation(customServiceHealthcheckValidation, ServiceHealthcheck{})
	if err := validate.Struct(newService); err != nil {
		return modifyValidateError(err)
	}

	if err := validateServiceBalanceType(newService.BalanceType); err != nil {
		return err
	}
	if err := validateServiceRoutingType(newService.RoutingType); err != nil {
		return err
	}
	if err := validateServiceProtocol(newService.Protocol); err != nil {
		return err
	}

	return nil
}

// func customPortValidationForCreateService(sl validator.StructLevel) {
// 	nbi := sl.Current().Interface().(NewServiceInfo)
// 	port, err := strconv.Atoi(nbi.ServicePort)
// 	if err != nil {
// 		sl.ReportError(nbi.ServicePort, "servicePort", "ServicePort", "port must be number", "")
// 	}
// 	if !(port > 0) || !(port < 20000) {
// 		sl.ReportError(nbi.ServicePort, "servicePort", "ServicePort", "port must gt=0 and lt=20000", "")
// 	}
// }
