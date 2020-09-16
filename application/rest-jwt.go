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
		validateIncomeError(validateError.Error(), tokenRequestUUID, ginContext, restAPI.balancerFacade.Logging)
		return
	}

	logChangeUUID(tokenRequestUUID, tokenRequest.ID, restAPI.balancerFacade.Logging)
	tokenRequestUUID = tokenRequest.ID

	if !restAPI.isValidUser(tokenRequest) {
		tokenResponseError := TokenResponseError{
			ID:    tokenRequestUUID,
			Error: "invalid login details for user " + tokenRequest.User,
		}
		ginContext.JSON(400, tokenResponseError)
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
	if err := validate.Struct(tokenRequest); err != nil {
		return modifyValidateError(err)
	}
	return nil
}

func (restAPI *RestAPIstruct) newTokens() (*TokenResponseOkay, error) {
	forNewToken := jwt_lib.New(jwt_lib.GetSigningMethod("HS256"))
	forNewToken.Claims = jwt_lib.MapClaims{
		"exp": time.Now().Add(restAPI.authorization.expireToken).Unix(),
	}

	forRefreshToken := jwt_lib.New(jwt_lib.GetSigningMethod("HS256"))
	forRefreshToken.Claims = jwt_lib.MapClaims{
		"exp": time.Now().Add(restAPI.authorization.expireTokenForRefresh).Unix(),
	}

	newToken, err := forNewToken.SignedString([]byte(restAPI.authorization.mainSecret))
	if err != nil {
		return nil, fmt.Errorf("Could not generate token")
	}

	refreshToken, err := forRefreshToken.SignedString([]byte(restAPI.authorization.mainSecretForRefresh))
	if err != nil {
		return nil, fmt.Errorf("Could not generate token")
	}

	return &TokenResponseOkay{
		AccessToken:  newToken,
		RefreshToken: refreshToken,
	}, nil
}

func (restAPI *RestAPIstruct) isValidUser(tokenRequest *TokenRequest) bool {
	user := strings.ToLower(tokenRequest.User)
	if password, ok := restAPI.authorization.credentials[user]; ok {
		if password == tokenRequest.Password {
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
// @Router /jwt/refresh-token [post]
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
