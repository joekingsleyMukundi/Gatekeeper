package api

import (
	"github.com/gin-gonic/gin"
	db "github.com/joekingsleyMukundi/Gatekeeper/db/sqlc"
	"github.com/joekingsleyMukundi/Gatekeeper/utils"
)

type Server struct {
	config utils.Config
	store  db.Store
	router *gin.Engine
}

func NewSever(config utils.Config, store db.Store) (*Server, error) {
	server := &Server{
		config: config,
		store:  store,
	}
	server.routerSetup()
	return server, nil
}

func (server *Server) routerSetup() {
	router := gin.Default()
	// TO DO : Create user suth apis
	server.router = router
}
func (server *Server) Start(address string) error {
	return server.router.Run(address)
}
func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
