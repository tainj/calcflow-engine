package config

import (
	"log"

	"github.com/caarlos0/env/v8"
	"github.com/joho/godotenv"
	"github.com/tainj/distributed_calculator2/internal/auth"
	"github.com/tainj/distributed_calculator2/pkg/db/cache"
	"github.com/tainj/distributed_calculator2/pkg/db/postgres"
	"github.com/tainj/distributed_calculator2/pkg/messaging/kafka"
)

type GRPCServer struct {
	RestPort int `env:"GRPC_REST_SERVER_PORT" env-default:"8080"`
	GRPCPort int `env:"GRPC_SERVER_PORT" env-default:"50051"`
}

type Config struct {
	Postgres postgres.Config
	Redis    cache.Config
	Grpc     GRPCServer
	Kafka    kafka.Config
	JWT      auth.Config
}

func LoadConfig() (*Config, error) {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	cfg := &Config{}

	if err := env.Parse(&cfg.Postgres); err != nil {
		return nil, err
	}

	if err := env.Parse(&cfg.Redis); err != nil {
		return nil, err
	}

	if err := env.Parse(&cfg.Grpc); err != nil {
		return nil, err
	}

	if err := env.Parse(&cfg.Kafka); err != nil {
		return nil, err
	}

	if err := env.Parse(&cfg.JWT); err != nil {
		return nil, err
	}

	return cfg, nil
}
