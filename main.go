package main

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/hibiken/asynq"
	"github.com/joekingsleyMukundi/Gatekeeper/api"
	db "github.com/joekingsleyMukundi/Gatekeeper/db/sqlc"
	"github.com/joekingsleyMukundi/Gatekeeper/internals"
	"github.com/joekingsleyMukundi/Gatekeeper/services/mail"
	"github.com/joekingsleyMukundi/Gatekeeper/utils"
	"github.com/joekingsleyMukundi/Gatekeeper/workers"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

func main() {
	config, err := utils.LoadConfig(".")
	if err != nil {
		log.Fatal("ERROR: cannot load config: ", err)
	}
	conn, err := sql.Open(config.DBdriver, config.DBsource)
	if err != nil {
		log.Fatal("ERROR: cannot connect to db: ", err)
	}
	store := db.NewStore(conn)
	redisOpt := asynq.RedisClientOpt{
		Addr: config.RedisAddress,
	}
	taskDistributor := workers.NewRedisTaskDistributor(redisOpt)
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddress,
		Password: "",
		DB:       0,
	})
	tokenManager := internals.NewTokenManager(rdb, context.Background(), config.RefreshTokenDuration+(24*time.Hour))
	go runTaskProcessor(redisOpt, store, config)
	runGinServer(config, store, taskDistributor, rdb, tokenManager)
}
func runGinServer(config utils.Config, store db.Store, taskDistributor workers.TaskDistributor, redisClient *redis.Client, tokenManager internals.Manager) {
	server, err := api.NewSever(config, store, taskDistributor, redisClient, tokenManager)
	if err != nil {
		log.Fatal("cannot connet to server:", err)
	}

	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("cannot connet to server:", err)
	}
}
func runTaskProcessor(
	redisOpt asynq.RedisClientOpt,
	store db.Store,
	config utils.Config,
) {
	mailer := mail.NewGmailSender(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword)
	taskProcessor := workers.NewRedisTaskProcessor(redisOpt, store, mailer)
	log.Println("start task processor")
	err := taskProcessor.Start()
	if err != nil {
		log.Fatal("failed to start task processor")
	}
}
