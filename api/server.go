package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	db "github.com/joekingsleyMukundi/Gatekeeper/db/sqlc"
	"github.com/joekingsleyMukundi/Gatekeeper/tokens"
	"github.com/joekingsleyMukundi/Gatekeeper/utils"
)

type Server struct {
	TokenMaker tokens.Maker
	config     utils.Config
	store      db.Store
	Router     *gin.Engine
}

func NewSever(config utils.Config, store db.Store) (*Server, error) {
	tokenMaker, err := tokens.NewJWTMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("ERROR: cannot create token: %s", err)
	}
	server := &Server{
		TokenMaker: tokenMaker,
		config:     config,
		store:      store,
	}
	server.routerSetup()
	return server, nil
}

func (server *Server) routerSetup() {
	router := gin.Default()
	// TO DO : Create user suth apis
	server.Router = router
}
func (server *Server) Start(address string) error {
	return server.Router.Run(address)
}
func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
