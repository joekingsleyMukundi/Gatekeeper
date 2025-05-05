package main

import (
	"database/sql"
	"log"

	"github.com/hibiken/asynq"
	"github.com/joekingsleyMukundi/Gatekeeper/api"
	db "github.com/joekingsleyMukundi/Gatekeeper/db/sqlc"
	"github.com/joekingsleyMukundi/Gatekeeper/utils"
	"github.com/joekingsleyMukundi/Gatekeeper/workers"
	_ "github.com/lib/pq"
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
	go runTaskProcessor(redisOpt, store)
	runGinServer(config, store, taskDistributor)
}
func runGinServer(config utils.Config, store db.Store, taskDistributor workers.TaskDistributor) {
	server, err := api.NewSever(config, store, taskDistributor)
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
) {
	taskProcessor := workers.NewRedisTaskProcessor(redisOpt, store)
	log.Println("start task processor")
	err := taskProcessor.Start()
	if err != nil {
		log.Fatal("failed to start task processor")
	}
}
