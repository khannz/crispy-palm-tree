package application

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

const modifyServiceRequestName = "modify service"

// modifyService godoc
// @tags Load balancer
// @Summary Modify service
// @Description Больше, чем балансировщик
// @Param addr path string true "IP"
// @Param port path uint true "Port"
// @Param incomeJSON body application.ModifyServiceInfo true "Expected json"
// @Accept json
// @Produce json
// @Success 200 {object} application.UniversalResponse "If all okay"
// @Failure 400 {object} application.UniversalResponse "Bad request"
// @Failure 500 {object} application.UniversalResponse "Internal error"
// @Router /service/{addr}/{port} [put]
// @Security ApiKeyAuth
func (restAPI *RestAPIstruct) modifyService(ginContext *gin.Context) {
	modifyServiceUUID := restAPI.balancerFacade.UUIDgenerator.NewUUID().UUID.String()
	restAPI.balancerFacade.Logging.WithFields(logrus.Fields{"event uuid": modifyServiceUUID}).Infof("got new %v request", modifyServiceRequestName)
	modifyService := &ModifyServiceInfo{}

	if err := ginContext.ShouldBindJSON(modifyService); err != nil {
		// TODO: log here
		unmarshallIncomeError(err.Error(),
			modifyServiceUUID,
			ginContext,
			restAPI.balancerFacade.Logging)
		return
	}
	ip := ginContext.Param("addr")
	port := ginContext.Param("port")
	modifyService.ServiceIP = ip
	modifyService.ServicePort = port

	if validateError := modifyService.validatemodifyService(); validateError != nil {
		// TODO: log here
		validateIncomeError(validateError.Error(), modifyServiceUUID, ginContext, restAPI.balancerFacade.Logging)
		return
	}

	nwbServiceInfo, err := restAPI.balancerFacade.ModifyService(modifyService,
		modifyServiceUUID)
	if err != nil {
		// TODO: log here
		uscaseFail(modifyServiceRequestName,
			err.Error(),
			modifyServiceUUID,
			ginContext,
			restAPI.balancerFacade.Logging)
		return
	}

	serviceInfo := convertDomainServiceInfoToRestUniversalResponse(nwbServiceInfo, true)

	// TODO: log here
	logRequestIsDone(modifyServiceRequestName, modifyServiceUUID, restAPI.balancerFacade.Logging)
	ginContext.JSON(http.StatusOK, serviceInfo)
}

func (modifyService *ModifyServiceInfo) validatemodifyService() error {
	validate := validator.New()
	validate.RegisterStructValidation(customPortValidationFormodifyService, ModifyServiceInfo{})
	validate.RegisterStructValidation(customPortServerApplicationValidation, ServerApplication{})
	validate.RegisterStructValidation(customServiceHealthcheckValidation, ServiceHealthcheck{})
	if err := validate.Struct(modifyService); err != nil {
		return modifyValidateError(err)
	}
	if err := validateServiceBalanceType(modifyService.BalanceType); err != nil {
		return err
	}
	if err := validateServiceRoutingType(modifyService.RoutingType); err != nil {
		return err
	}
	if err := validateServiceProtocol(modifyService.Protocol); err != nil {
		return err
	}
	return nil
}

func customPortValidationFormodifyService(sl validator.StructLevel) {
	nbi := sl.Current().Interface().(ModifyServiceInfo)
	port, err := strconv.Atoi(nbi.ServicePort)
	if err != nil {
		sl.ReportError(nbi.ServicePort, "servicePort", "ServicePort", "port must be number", "")
	}
	if !(port > 0) || !(port < 20000) {
		sl.ReportError(nbi.ServicePort, "servicePort", "ServicePort", "port must gt=0 and lt=20000", "")
	}
}
