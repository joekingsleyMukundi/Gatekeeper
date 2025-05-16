package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joekingsleyMukundi/Gatekeeper/middlewares"
	"github.com/joekingsleyMukundi/Gatekeeper/tokens"
	"github.com/joekingsleyMukundi/Gatekeeper/utils"
	"github.com/stretchr/testify/require"
)

const (
	authorizationHeaderKey  = "authorization"
	authorizationTYpeBearor = "bearer"
	authorizationPayloadKey = "authorization_payload"
)

func addAuthorization(
	t *testing.T,
	request *http.Request,
	tokenMaker tokens.Maker,
	authorizationType string,
	username string,
	duration time.Duration,
) {
	token, payload, err := tokenMaker.CreateToken(username, duration)
	require.NoError(t, err)
	require.NotEmpty(t, payload)
	authorizationHeader := fmt.Sprintf("%s %s", authorizationType, token)
	request.Header.Set(authorizationHeaderKey, authorizationHeader)
}
func TestAuthMiddleware(t *testing.T) {
	username := utils.RandomUsername()
	testCases := []struct {
		name          string
		setupAuth     func(t *testing.T, request *http.Request, tMaker tokens.Maker)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker tokens.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTYpeBearor, username, time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "NoAuthorization",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker tokens.Maker) {
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "UnsupportedAuthorization",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker tokens.Maker) {
				addAuthorization(t, request, tokenMaker, "unsupported", username, time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "InvalidAuthorizationFormat",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker tokens.Maker) {
				addAuthorization(t, request, tokenMaker, "", username, time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "ExpiredToken",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker tokens.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTYpeBearor, username, -time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			server := NewTestServer(t, nil, nil, nil)
			authPath := "/auth"
			server.Router.GET(
				authPath,
				middlewares.AuthMiddleware(server.TokenMaker),
				func(ctx *gin.Context) {
					ctx.JSON(http.StatusOK, gin.H{})
				},
			)
			recorder := httptest.NewRecorder()
			request, err := http.NewRequest(http.MethodGet, authPath, nil)
			require.NoError(t, err)
			tc.setupAuth(t, request, server.TokenMaker)
			server.Router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}
