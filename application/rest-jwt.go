package application

import (
	"fmt"
	"strings"
	"time"

	jwt_lib "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

const (
	requestTokenName = "token request"
	tokenRefreshName = "token refresh"
)

// login godoc
// @tags auth
// @Summary Create jwt
// @Description Make jwt easier ;)
// @Param incomeJSON body application.TokenRequest true "Expected json"
// @Accept json
// @Produce json
// @Success 200 {object} application.LoginResponseOkay "If all okay"
// @Failure 400 {object} application.LoginResponseError "Bad request"
// @Failure 500 {object} application.LoginResponseError "Internal error"
// @Router /login [post]
func (restAPI *RestAPIstruct) loginRequest(ginContext *gin.Context) {
	loginRequestUUID := restAPI.balancerFacade.UUIDgenerator.NewUUID().UUID.String()
	logNewRequest(requestTokenName, loginRequestUUID, restAPI.balancerFacade.Logging)

	loginRequest := &LoginRequest{}

	if err := ginContext.ShouldBindJSON(loginRequest); err != nil {
		unmarshallIncomeError(err.Error(),
			loginRequestUUID,
			ginContext,
			restAPI.balancerFacade.Logging)
		return
	}

	if validateError := loginRequest.validateLoginRequest(); validateError != nil {
		validateIncomeError(validateError.Error(), loginRequestUUID, ginContext, restAPI.balancerFacade.Logging)
		return
	}

	if !restAPI.isValidUser(loginRequest) {
		tokenResponseError := LoginResponseError{
			Error: "invalid login details for user " + loginRequest.User,
		}
		ginContext.JSON(400, tokenResponseError)
		return
	}

	token, refreshToken, err := restAPI.newTokens()
	if err != nil {
		tokenResponseError := LoginResponseError{
			Error: err.Error(),
		}
		ginContext.JSON(500, tokenResponseError)
		return
	}

	ginContext.SetCookie("JWT", refreshToken, 86400000, "/", restAPI.ip, false, true)

	logRequestIsDone(requestTokenName, loginRequestUUID, restAPI.balancerFacade.Logging)

	ginContext.JSON(200, LoginResponseOkay{AccessToken: token})
}

func (loginRequest *LoginRequest) validateLoginRequest() error {
	validate := validator.New()
	if err := validate.Struct(loginRequest); err != nil {
		return modifyValidateError(err)
	}
	return nil
}

func (restAPI *RestAPIstruct) newTokens() (string, string, error) {
	forNewToken := jwt_lib.New(jwt_lib.GetSigningMethod("HS256"))
	forNewToken.Claims = jwt_lib.MapClaims{
		"exp": time.Now().Add(restAPI.authorization.expireToken).Unix(),
	}

	forRefreshToken := jwt_lib.New(jwt_lib.GetSigningMethod("HS256"))
	forRefreshToken.Claims = jwt_lib.MapClaims{
		// "exp": time.Now().Add(restAPI.authorization.expireTokenForRefresh).Unix(),
	}

	newToken, err := forNewToken.SignedString([]byte(restAPI.authorization.mainSecret))
	if err != nil {
		return "", "", fmt.Errorf("can't not generate new token")
	}

	refreshToken, err := forRefreshToken.SignedString([]byte(restAPI.authorization.mainSecretForRefresh))
	if err != nil {
		return "", "", fmt.Errorf("can't not generate refresh token")
	}

	return newToken, refreshToken, nil
}

func (restAPI *RestAPIstruct) isValidUser(loginRequest *LoginRequest) bool {
	user := strings.ToLower(loginRequest.User)
	if password, ok := restAPI.authorization.credentials[user]; ok {
		if password == loginRequest.Password {
			return true
		}
	}
	return false
}

// tokenRefresh godoc
// @tags auth
// @Summary Refresh jwt
// @Description Make jwt easier ;)
// @Produce json
// @Success 200 {object} application.LoginResponseOkay "If all okay"
// @Failure 400 {object} application.LoginResponseError "Bad request"
// @Failure 500 {object} application.LoginResponseError "Internal error"
// @Router /token [get]
// @Security ApiKeyAuth
func (restAPI *RestAPIstruct) tokenRefresh(ginContext *gin.Context) {
	tokenRefreshUUID := restAPI.balancerFacade.UUIDgenerator.NewUUID().UUID.String()
	logNewRequest(tokenRefreshName, tokenRefreshUUID, restAPI.balancerFacade.Logging)

	// cookie, err := c.Cookie("JWT") // TODO: fail

	token, refreshToken, err := restAPI.newTokens()
	if err != nil {
		tokenResponseError := LoginResponseError{
			Error: err.Error(),
		}
		ginContext.JSON(500, tokenResponseError)
		return
	}

	ginContext.SetCookie("JWT", refreshToken, 86400000, "/", restAPI.ip, false, true)

	logRequestIsDone(tokenRefreshName, tokenRefreshUUID, restAPI.balancerFacade.Logging)

	ginContext.JSON(200, LoginResponseOkay{AccessToken: token})
}
