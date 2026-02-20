// cmd/worker/main.go
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	repo "github.com/tainj/distributed_calculator2/internal/repository"
	"github.com/tainj/distributed_calculator2/internal/valueprovider"
	"github.com/tainj/distributed_calculator2/internal/worker"
	"github.com/tainj/distributed_calculator2/pkg/config"
	"github.com/tainj/distributed_calculator2/pkg/db/cache"
	"github.com/tainj/distributed_calculator2/pkg/db/postgres"
	"github.com/tainj/distributed_calculator2/pkg/logger"
	"github.com/tainj/distributed_calculator2/pkg/messaging/kafka"
)

func main() {
	ctx := context.Background()
	mainLogger := logger.New("worker")

	// Worker port - from env
	port := os.Getenv("WORKER_PORT")
	if port == "" {
		port = "8081"
	}
	workerLogger := mainLogger.With("component", "Worker", "port", port)

	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Println("failed to load config:", err)
		os.Exit(1)
	}

	redis := cache.New(cfg.Redis, mainLogger)
	db, err := postgres.New(cfg.Postgres)
	if err != nil {
		mainLogger.Error(ctx, "failed to connect to postgres", "error", err)
		os.Exit(1)
	}

	factory := repo.NewRepositoryFactory(db, redis, mainLogger)
	exampleRepo := factory.CreateExampleRepository()
	variableRepo := factory.CreateVariableRepository()

	kafkaLogger := mainLogger.With("component", "KafkaConsumer")
	kafkaQueue, err := kafka.NewKafkaQueue(cfg.Kafka, kafkaLogger)
	if err != nil {
		mainLogger.Error(ctx, "failed to init kafka", "error", err)
		os.Exit(1)
	}

	valueProvider := valueprovider.NewRedisValueProvider(redis)

	w := worker.NewWorker(exampleRepo, variableRepo, kafkaQueue, valueProvider, workerLogger, port)

	// Run
	go w.Start()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	w.Stop()
}
