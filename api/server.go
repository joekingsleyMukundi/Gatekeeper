package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	db "github.com/joekingsleyMukundi/Gatekeeper/db/sqlc"
	"github.com/joekingsleyMukundi/Gatekeeper/tokens"
	"github.com/joekingsleyMukundi/Gatekeeper/utils"
	"github.com/joekingsleyMukundi/Gatekeeper/workers"
)

type Server struct {
	TokenMaker      tokens.Maker
	config          utils.Config
	store           db.Store
	Router          *gin.Engine
	taskDistributor workers.TaskDistributor
}

func NewSever(config utils.Config, store db.Store, taskDistributor workers.TaskDistributor) (*Server, error) {
	tokenMaker, err := tokens.NewJWTMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("ERROR: cannot create token: %s", err)
	}
	server := &Server{
		TokenMaker:      tokenMaker,
		config:          config,
		store:           store,
		taskDistributor: taskDistributor,
	}
	server.routerSetup()
	return server, nil
}

func (server *Server) routerSetup() {
	router := gin.Default()
	// TO DO : Create user suth apis
	router.POST("/api/v1/auth/register", server.createUser)
	router.POST("/api/v1/auth/login", server.loginUser)
	router.POST("/api/v1/auth/token/refresh", server.renewAccessToken)
	router.POST("/api/v1/auth/password/forgot", server.forgotPassword)
	router.PATCH("/api/v1/auth/password/reset/:token", server.resetPassword)
	router.GET("/api/v1/auth/email/verify/:token", server.validateEmail)
	server.Router = router
}
func (server *Server) Start(address string) error {
	return server.Router.Run(address)
}
func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
