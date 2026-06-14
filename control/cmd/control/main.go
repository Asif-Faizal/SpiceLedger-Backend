package main

import (
	"log"

	"github.com/Asif-Faizal/SpiceLedger-Backend/control"
	"github.com/Asif-Faizal/SpiceLedger-Backend/util"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// 1. Load Configuration
	config := util.LoadConfig()

	// 2. Initialize Logger
	logger := util.NewLogger(config.LogLevel)

	// 3. Initialize Repository
	repo, err := control.NewMysqlRepository(config.DSN(), logger)
	if err != nil {
		log.Fatalf("could not create database repository: %v", err)
	}
	defer repo.Close()

	// 4. Initialize Service
	accountService := control.NewAccountService(
		repo,
		config.JWTSecret,
		config.AccessTokenDuration,
		config.RefreshTokenDuration,
	)

	// 5. Start gRPC Server
	if err := control.ListenGrpcServer(accountService, logger, config); err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
}
