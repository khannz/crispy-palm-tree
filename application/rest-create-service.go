package application

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

const createServiceRequestName = "create service"

// createService godoc
// @tags load balancer
// @Summary Create service
// @Description Больше, чем балансировщик
// @Param addr path string true "IP"
// @Param port path uint true "Port"
// @Param incomeJSON body application.NewServiceInfo true "Expected json"
// @Accept json
// @Produce json
// @Success 200 {object} application.UniversalResponse "If all okay"
// @Failure 400 {object} application.UniversalResponse "Bad request"
// @Failure 500 {object} application.UniversalResponse "Internal error"
// @Router /service/{addr}/{port} [post]
// // // @Security ApiKeyAuth
func (restAPI *RestAPIstruct) createService(ginContext *gin.Context) {
	createServiceID := restAPI.balancerFacade.IDgenerator.NewID()
	restAPI.balancerFacade.Logging.WithFields(logrus.Fields{"event id": createServiceID}).Infof("got new %v request", createServiceRequestName)
	createService := &NewServiceInfo{}

	if err := ginContext.ShouldBindJSON(createService); err != nil {
		// TODO: log here
		unmarshallIncomeError(err.Error(),
			createServiceID,
			ginContext,
			restAPI.balancerFacade.Logging)
		return
	}
	ip := ginContext.Param("addr")
	port := ginContext.Param("port")
	createService.ServiceIP = ip
	createService.ServicePort = port

	if validateError := createService.validateCreateService(); validateError != nil {
		// TODO: log and response here
		validateIncomeError(validateError.Error(), createServiceID, ginContext, restAPI.balancerFacade.Logging)
		return
	}

	nwbServiceInfo, err := restAPI.balancerFacade.CreateService(createService,
		createServiceID)
	if err != nil {
		restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
			"event id": createServiceID,
		}).Errorf("can't %v, got error: %v", createServiceRequestName, err)
		rError := &UniversalResponse{
			ID:                       createServiceID,
			JobCompletedSuccessfully: false,
			ExtraInfo:                "got internal error: %b" + err.Error(),
		}
		ginContext.JSON(http.StatusInternalServerError, rError)
		return
	}

	serviceInfo := convertDomainServiceInfoToRestUniversalResponse(nwbServiceInfo, true)

	restAPI.balancerFacade.Logging.WithFields(logrus.Fields{
		"event id": createServiceID,
	}).Infof("request %v done", createServiceRequestName)

	ginContext.JSON(http.StatusOK, serviceInfo)
}

func (createService *NewServiceInfo) validateCreateService() error {
	validate := validator.New()
	validate.RegisterStructValidation(customPortValidationForCreateService, NewServiceInfo{})
	validate.RegisterStructValidation(customPortServerApplicationValidation, ServerApplication{})
	validate.RegisterStructValidation(customServiceHealthcheckValidation, ServiceHealthcheck{})
	if err := validate.Struct(createService); err != nil {
		return modifyValidateError(err)
	}

	if err := validateServiceBalanceType(createService.BalanceType); err != nil {
		return err
	}
	if err := validateServiceRoutingType(createService.RoutingType); err != nil {
		return err
	}
	if err := validateServiceProtocol(createService.Protocol); err != nil {
		return err
	}

	return nil
}

func customPortValidationForCreateService(sl validator.StructLevel) {
	nbi := sl.Current().Interface().(NewServiceInfo)
	port, err := strconv.Atoi(nbi.ServicePort)
	if err != nil {
		sl.ReportError(nbi.ServicePort, "servicePort", "ServicePort", "port must be number", "")
	}
	if !(port > 0) || !(port < 20000) {
		sl.ReportError(nbi.ServicePort, "servicePort", "ServicePort", "port must gt=0 and lt=20000", "")
	}
}
