package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/tainj/distributed_calculator2/internal/auth"
	repo "github.com/tainj/distributed_calculator2/internal/repository"
	service "github.com/tainj/distributed_calculator2/internal/service"
	"github.com/tainj/distributed_calculator2/internal/transport/grpc"
	"github.com/tainj/distributed_calculator2/pkg/config"
	"github.com/tainj/distributed_calculator2/pkg/db/cache"
	"github.com/tainj/distributed_calculator2/pkg/db/postgres"
	"github.com/tainj/distributed_calculator2/pkg/logger"
	"github.com/tainj/distributed_calculator2/pkg/messaging/kafka"
)

const (
	serviceName = "distributed_calculator"
)

func main() {
	// Basic context
	ctx := context.Background()

	// Initialize loggers
	mainLogger := logger.New(serviceName)

	// Extend logger for components
	kafkaLogger := mainLogger.With("component", "KafkaConsumer")
	// workerLogger := mainLogger.With("component", "Worker", "worker_id", "1")
	// httpLogger := mainLogger.With("handler", "CalculateHandler")
	ctx = context.WithValue(ctx, logger.LoggerKey, mainLogger)

	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if cfg == nil {
		mainLogger.Error(ctx, "failed to load config")
		os.Exit(1)
	}

	// 1. Database
	db, err := postgres.New(cfg.Postgres)
	if err != nil {
		fmt.Println(err)
		mainLogger.Error(ctx, "failed to init postgres", "error", err)
		os.Exit(1)
	}

	// 2. Cache (redis)
	redis := cache.New(cfg.Redis, mainLogger)
	fmt.Println(redis.Client.Ping(ctx)) // Check connection

	// 3. Repository factory
	factory := repo.NewRepositoryFactory(db, redis, mainLogger)

	// 4. JWT service - needed for auth middleware
	jwtService := auth.NewJWTService(cfg.JWT)

	// 5. Kafka - task queue
	kafkaQueue, err := kafka.NewKafkaQueue(cfg.Kafka, kafkaLogger)
	if err != nil {
		mainLogger.Error(ctx, "failed to init kafka: "+err.Error())
		os.Exit(1)
	}

	// 6. ValueProvider - for retrieving variables from redis
	// valueProvider := valueprovider.NewRedisValueProvider(redis)

	// 7. Repositories
	// variableRepo := factory.CreateVariableRepository() // for saving results
	exampleRepo := factory.CreateExampleRepository() // for saving expressions
	userRepo := factory.CreateUserRepository()

	// 8. Calculator service
	srv := service.NewCalculatorService(userRepo, exampleRepo, jwtService, kafkaQueue, mainLogger)

	// 9. Worker - processes tasks from kafka
	// worker := worker.NewWorker(exampleRepo, variableRepo, kafkaQueue, valueProvider, workerLogger)
	// go worker.Start() // in separate goroutine

	// 10. gRPC server (gRPC + REST via gateway)
	grpcServer, err := grpc.New(ctx, cfg.Grpc.GRPCPort, cfg.Grpc.RestPort, srv, jwtService)
	if err != nil {
		mainLogger.Error(ctx, err.Error())
		return
	}

	// graceful shutdown
	graceCh := make(chan os.Signal, 1)
	signal.Notify(graceCh, syscall.SIGINT, syscall.SIGTERM)

	// Start server asynchronously
	go func() {
		if err := grpcServer.Start(ctx); err != nil {
			mainLogger.Error(ctx, err.Error())
		}
	}()

	// Wait for stop signal
	<-graceCh

	// Stop
	if err := grpcServer.Stop(ctx); err != nil {
		mainLogger.Error(ctx, err.Error())
	}
	mainLogger.Info(ctx, "server stopped")
}
