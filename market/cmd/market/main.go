package main

import (
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"

	"github.com/Asif-Faizal/SpiceLedger-Backend/market"
	"github.com/Asif-Faizal/SpiceLedger-Backend/util"
)

func main() {
	// 1. Load Configuration
	config := util.LoadConfig()

	// 2. Initialize Logger
	logger := util.NewLogger(config.LogLevel)

	// 3. Initialize Repository
	dbURL := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		config.DBUser, config.DBPass, config.DBHost, config.DBPort, config.DBName)
	repo, err := market.NewMysqlRepository(dbURL, logger)
	if err != nil {
		log.Fatalf("could not create database repository: %v", err)
	}
	defer repo.Close()

	// 4. Initialize Service
	marketService := market.NewMarketService(repo, logger)

	// 5. Start gRPC Server
	if err := market.ListenGrpcServer(marketService, logger, config); err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
}
