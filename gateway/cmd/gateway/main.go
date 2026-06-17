package main

import (
	"fmt"

	"github.com/Asif-Faizal/SpiceLedger-Backend/gateway"
	"github.com/Asif-Faizal/SpiceLedger-Backend/internal/platform"
	"github.com/Asif-Faizal/SpiceLedger-Backend/util"
)

func main() {
	cfg := util.LoadConfig()
	logger := util.NewLogger(cfg.LogLevel)

	deps, err := gateway.NewDependencies(cfg, logger)
	if err != nil {
		logger.Service().Fatal().Err(err).Msg("gateway initialization failed")
	}
	defer func() {
		if err := deps.Close(); err != nil {
			logger.Service().Error().Err(err).Msg("gateway shutdown cleanup error")
		}
	}()

	if err := platform.RunHTTP(platform.HTTPConfig{
		Name:    "gateway",
		Addr:    fmt.Sprintf(":%d", cfg.ProxyPort),
		Handler: gateway.NewHandler(deps),
	}, logger); err != nil {
		logger.Service().Fatal().Err(err).Msg("gateway terminated with error")
	}
}
