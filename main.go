package main

import (
	"database/sql"
	"log"

	"github.com/joekingsleyMukundi/Gatekeeper/api"
	db "github.com/joekingsleyMukundi/Gatekeeper/db/sqlc"
	"github.com/joekingsleyMukundi/Gatekeeper/utils"
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
	server, err := api.NewSever(config, store)
	if err != nil {
		log.Fatal("cannot creare server:", err)
	}
	server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("cannot connet to server:", err)
	}
}
