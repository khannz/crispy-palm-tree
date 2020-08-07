package application

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

const getServicesRequestName = "get services"

// getServices godoc
// @tags Network balance services
// @Summary Get all services
// @Description Make network balance service easier ;)
// @Param incomeJSON body application.GetAllServicesRequest true "Expected json"
// @Accept json
// @Produce json
// @Success 200 {object} application.GetAllServicesResponse "If all okay"
// @Failure 400 {object} application.GetAllServicesResponse "Bad request"
// @Failure 500 {object} application.GetAllServicesResponse "Internal error"
// @Router /service/get-services [post]
func (restAPI *RestAPIstruct) getServices(ginContext *gin.Context) {
	getServicesRequestUUID := restAPI.balancerFacade.UUIDgenerator.NewUUID().UUID.String()
	logNewRequest(getServicesRequestName, getServicesRequestUUID, restAPI.balancerFacade.Logging)

	newGetServicesRequest := &GetAllServicesRequest{}

	if err := ginContext.ShouldBindJSON(newGetServicesRequest); err != nil {
		unmarshallIncomeError(err.Error(),
			getServicesRequestUUID,
			ginContext,
			restAPI.balancerFacade.Logging)
		return
	}

	validateError := newGetServicesRequest.validateGetServicesRequest()
	if validateError != nil {
		stringValidateError := errorsValidateToString(validateError)
		validateIncomeError(stringValidateError, getServicesRequestUUID, ginContext, restAPI.balancerFacade.Logging)
		return
	}

	logChangeUUID(getServicesRequestUUID, newGetServicesRequest.ID, restAPI.balancerFacade.Logging)
	getServicesRequestUUID = newGetServicesRequest.ID

	nwbServices, err := restAPI.balancerFacade.GetServices(getServicesRequestUUID)
	if err != nil {
		uscaseFail(getServicesRequestName,
			err.Error(),
			getServicesRequestUUID,
			ginContext,
			restAPI.balancerFacade.Logging)
		return
	}

	logRequestIsDone(getServicesRequestName, getServicesRequestUUID, restAPI.balancerFacade.Logging)
	allServices := convertDomainServicesInfoToRestUniversalResponseWithState(nwbServices, true)

	var extraInfo string
	if len(allServices) == 0 {
		extraInfo = "No services here"
	}
	getServicesResponse := GetAllServicesResponse{
		ID:                       getServicesRequestUUID,
		JobCompletedSuccessfully: true,
		AllServices:              allServices,
		ExtraInfo:                extraInfo,
	}

	ginContext.JSON(http.StatusOK, gin.H{"data": getServicesResponse})
}

func (getAllServicesRequest *GetAllServicesRequest) validateGetServicesRequest() error {
	validate := validator.New()
	err := validate.Struct(getAllServicesRequest)
	if err != nil {
		return err
	}
	return nil
}
