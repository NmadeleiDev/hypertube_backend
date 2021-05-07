package main

import (
	"hypertube_storage/db"
	"hypertube_storage/parser/env"
	"hypertube_storage/server"
	"torrentClient/eraser"
)

func main() {
	InitLog()
	db.GetLoadedFilesManager().InitConnection(env.GetParser().GetPostgresDbDsn())
	db.GetLoadedFilesManager().InitTables()

	db.GetLoadedStateDb().InitConnection()

	defer func() {
		db.GetLoadedFilesManager().CloseConnection()
		db.GetLoadedStateDb().CloseConnection()
	}()

	go eraser.GetEraser().StartCheckingForRecords()

	server.Start()
}

