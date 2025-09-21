package main

import (
	"fmt"

	"github.com/MussaShaukenov/go-concurrency/internal/config"
	"github.com/MussaShaukenov/go-concurrency/internal/network"
	"github.com/MussaShaukenov/go-concurrency/internal/storage/engine"
	"github.com/MussaShaukenov/go-concurrency/internal/utils/logger"
)

const configPath = "../../config.yaml"

func main() {
	cfg, err := config.New(configPath)
	if err != nil {
		panic(err)
	}

	log, err := logger.NewLogger(cfg)
	if err != nil {
		panic(fmt.Sprintf("failed to create logger: %v", err))
	}

	storageEngine := engine.NewEngine(log)
	server, err := network.NewTCPServer(log, cfg.Network, storageEngine)
	if err != nil {
		log.Fatalf("failed to create server: %w", err)
	}

	if err := server.Start(); err != nil {
		log.Fatalf("failed to start server: %w", err)
	}
}
