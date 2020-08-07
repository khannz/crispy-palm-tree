package application

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	jwt_lib "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

const (
	requestTokenName = "token request"
	refreshTokenName = "token refresh"
)

// newToken godoc
// @tags jwt
// @Summary Create jwt
// @Description Make jwt easier ;)
// @Param incomeJSON body application.TokenRequest true "Expected json"
// @Accept json
// @Produce json
// @Success 200 {object} application.TokenResponseOkay "If all okay"
// @Failure 400 {object} application.TokenResponseError "Bad request"
// @Failure 500 {object} application.TokenResponseError "Internal error"
// @Router /jwt/request-token [post]
func (restAPI *RestAPIstruct) tokenRequest(ginContext *gin.Context) {
	tokenRequestUUID := restAPI.balancerFacade.UUIDgenerator.NewUUID().UUID.String()
	logNewRequest(requestTokenName, tokenRequestUUID, restAPI.balancerFacade.Logging)

	tokenRequest := &TokenRequest{}

	if err := ginContext.ShouldBindJSON(tokenRequest); err != nil {
		unmarshallIncomeError(err.Error(),
			tokenRequestUUID,
			ginContext,
			restAPI.balancerFacade.Logging)
		return
	}

	if validateError := tokenRequest.validateTokenRequest(); validateError != nil {
		stringValidateError := errorsValidateToString(validateError)
		validateIncomeError(stringValidateError, tokenRequestUUID, ginContext, restAPI.balancerFacade.Logging)
		return
	}

	logChangeUUID(tokenRequestUUID, tokenRequest.ID, restAPI.balancerFacade.Logging)
	tokenRequestUUID = tokenRequest.ID

	if !restAPI.isValidUser(tokenRequest) {
		tokenResponseError := TokenResponseError{
			ID:    tokenRequestUUID,
			Error: "invalid login details",
		}
		ginContext.JSON(500, tokenResponseError)
		return
	}

	tokenResponseOkay, err := restAPI.newTokens()
	if err != nil {
		tokenResponseError := TokenResponseError{
			ID:    tokenRequestUUID,
			Error: err.Error(),
		}
		ginContext.JSON(500, tokenResponseError)
		return
	}
	tokenResponseOkay.ID = tokenRequestUUID

	logRequestIsDone(requestTokenName, tokenRequestUUID, restAPI.balancerFacade.Logging)

	ginContext.JSON(200, tokenResponseOkay)
}

func (tokenRequest *TokenRequest) validateTokenRequest() error {
	validate := validator.New()
	validate.RegisterStructValidation(customPortValidationForTokenRequest, TokenRequest{})
	return nil
}

func customPortValidationForTokenRequest(sl validator.StructLevel) {
	nbi := sl.Current().Interface().(NewServiceInfo)
	port, err := strconv.Atoi(nbi.ServicePort)
	if err != nil {
		sl.ReportError(nbi.ServicePort, "servicePort", "ServicePort", "port must be number", "")
	}
	if !(port > 0) || !(port < 20000) {
		sl.ReportError(nbi.ServicePort, "servicePort", "ServicePort", "port must gt=0 and lt=20000", "")
	}
}

func (restAPI *RestAPIstruct) newTokens() (*TokenResponseOkay, error) {
	forNewToken := jwt_lib.New(jwt_lib.GetSigningMethod("HS256"))
	forNewToken.Claims = jwt_lib.MapClaims{
		"exp": time.Now().Add(time.Hour * 10).Unix(),
	}

	forRefreshToken := jwt_lib.New(jwt_lib.GetSigningMethod("HS256"))
	forRefreshToken.Claims = jwt_lib.MapClaims{
		"exp": time.Now().Add(time.Hour * 96).Unix(),
	}

	newToken, err := forNewToken.SignedString([]byte(restAPI.authorization.mainSecret))
	if err != nil {
		return nil, fmt.Errorf("Could not generate token")
	}

	refreshToken, err := forRefreshToken.SignedString([]byte(restAPI.authorization.mainRefreshSecret))
	if err != nil {
		return nil, fmt.Errorf("Could not generate token")
	}

	return &TokenResponseOkay{
		AccessToken:  newToken,
		RefreshToken: refreshToken,
	}, nil
}

func (restAPI *RestAPIstruct) isValidUser(tokenRequest *TokenRequest) bool {
	for _, user := range restAPI.authorization.users {
		if user.login == strings.ToLower(tokenRequest.User) && user.password == tokenRequest.Password {
			return true
		}
	}
	return false
}

// refreshToken godoc
// @tags jwt
// @Summary refresh jwt
// @Description Make jwt easier ;)
// @Produce json
// @Success 200 {object} application.TokenResponseOkay "If all okay"
// @Failure 400 {object} application.TokenResponseError "Bad request"
// @Failure 500 {object} application.TokenResponseError "Internal error"
// @Router /jwt/refresh-token [get]
// @Security ApiKeyAuth
func (restAPI *RestAPIstruct) tokenRefresh(ginContext *gin.Context) {
	tokenRequestUUID := restAPI.balancerFacade.UUIDgenerator.NewUUID().UUID.String()
	logNewRequest(refreshTokenName, tokenRequestUUID, restAPI.balancerFacade.Logging)

	tokenResponseOkay, err := restAPI.newTokens()
	if err != nil {
		tokenResponseError := TokenResponseError{
			ID:    tokenRequestUUID,
			Error: err.Error(),
		}
		ginContext.JSON(500, tokenResponseError)
		return
	}
	tokenResponseOkay.ID = tokenRequestUUID

	logRequestIsDone(refreshTokenName, tokenRequestUUID, restAPI.balancerFacade.Logging)

	ginContext.JSON(200, tokenResponseOkay)
}
