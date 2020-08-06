package application

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

const getServicesRequestName = "get services"

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
// @Router /get-services [post]
func (restAPI *RestAPIstruct) getServices(ginContext *gin.Context) {
	getServicesRequestUUID := restAPI.balancerFacade.UUIDgenerator.NewUUID().UUID.String()
	logNewRequest(getServicesRequestName, getServicesRequestUUID, restAPI.balancerFacade.Logging)

	var err error
	bytesFromBuf := readIncomeBytes(ginContext.Request)

	newGetServicesRequest := &GetAllServicesRequest{}

	err = json.Unmarshal(bytesFromBuf, newGetServicesRequest)
	if err != nil {
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
