package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
	"github.com/nokkvi/simplebank/api"
	db "github.com/nokkvi/simplebank/db/sqlc"
	"github.com/nokkvi/simplebank/util"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}
	conn, err := sql.Open(config.DBDrvier, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to database:", err)
	}

	store := db.NewStore(conn)
	server := api.NewServer(store)

	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("cannot start server:", err)
	}

}
