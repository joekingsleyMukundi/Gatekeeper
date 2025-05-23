package api

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	db "github.com/joekingsleyMukundi/Gatekeeper/db/sqlc"
	"github.com/joekingsleyMukundi/Gatekeeper/internals"
	"github.com/joekingsleyMukundi/Gatekeeper/middlewares"
	"github.com/joekingsleyMukundi/Gatekeeper/tokens"
	"github.com/joekingsleyMukundi/Gatekeeper/utils"
	"github.com/joekingsleyMukundi/Gatekeeper/workers"
	"github.com/redis/go-redis/v9"
)

type Server struct {
	TokenMaker      tokens.Maker
	config          utils.Config
	store           db.Store
	Router          *gin.Engine
	taskDistributor workers.TaskDistributor
	redisClient     *redis.Client
	//tokenManager    internals.Manager
}

func NewSever(config utils.Config, store db.Store, taskDistributor workers.TaskDistributor, redisClient *redis.Client) (*Server, error) {
	tokenMaker, err := tokens.NewJWTMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("ERROR: cannot create token: %s", err)
	}
	server := &Server{
		TokenMaker:      tokenMaker,
		config:          config,
		store:           store,
		taskDistributor: taskDistributor,
		redisClient:     redisClient,
		// tokenManager:    tokenManager,
	}
	server.routerSetup()
	return server, nil
}

func (server *Server) routerSetup() {
	_, err := server.redisClient.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}
	router := gin.Default()
	iprateLimiter := internals.NewSlidingWindowLimiter(server.redisClient, 5, 15*time.Minute)
	emailrateLimiter := internals.NewSlidingWindowLimiter(server.redisClient, 3, 30*time.Minute)

	router.POST("/api/v1/auth/register", server.createUser)
	router.POST("/api/v1/auth/login", server.loginUser)
	router.POST("/api/v1/auth/token/refresh", server.renewAccessToken)
	router.POST("/api/v1/auth/password/forgot", middlewares.ForgotPasswordRateLimitMiddleware(iprateLimiter, emailrateLimiter), server.forgotPassword)
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
