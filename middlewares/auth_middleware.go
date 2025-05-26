package middlewares

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joekingsleyMukundi/Gatekeeper/internals"
	"github.com/joekingsleyMukundi/Gatekeeper/tokens"
)

const (
	authorizationHeaderKey  = "authorization"
	authorizationTYpeBearor = "bearer"
	authorizationPayloadKey = "authorization_payload"
)

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
func AuthMiddleware(tMaker tokens.Maker, tm internals.Manager) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorizationHeader := ctx.GetHeader(authorizationHeaderKey)
		if len(authorizationHeader) == 0 {
			err := errors.New("ERROR: Authorization header was not provided")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}
		authField := strings.Fields(authorizationHeader)
		if len(authField) < 2 {
			err := errors.New("ERROR: Inavalid authorization format")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}
		authorizaionType := strings.ToLower(authField[0])
		if authorizaionType != authorizationTYpeBearor {
			err := fmt.Errorf("ERROR: Unsupported authorization header, %s", authorizaionType)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}
		accessToken := authField[1]
		isRevoked, err := tm.IsAccessTokenRevoked(accessToken)
		if err != nil {
			err := fmt.Errorf("ERROR: Issue in verifying accesstoken validity, %s", err)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		if isRevoked {
			err := fmt.Errorf("ERROR: Accesstoken revoked, please loging again")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}
		payload, err := tMaker.VerifyToken(accessToken)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}
		ctx.Set(authorizationPayloadKey, payload)
	}
}
