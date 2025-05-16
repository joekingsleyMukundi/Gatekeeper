package api

import (
	"testing"
	"time"

	db "github.com/joekingsleyMukundi/Gatekeeper/db/sqlc"
	"github.com/joekingsleyMukundi/Gatekeeper/utils"
	"github.com/joekingsleyMukundi/Gatekeeper/workers"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func NewTestServer(t *testing.T, store db.Store, taskDistributor workers.TaskDistributor, redisClient *redis.Client) *Server {
	config := utils.Config{
		TokenSymmetricKey:   utils.RandomString(32),
		AccessTokenDuration: time.Minute,
	}

	server, err := NewSever(config, store, taskDistributor, redisClient)
	require.NoError(t, err)
	return server
}
