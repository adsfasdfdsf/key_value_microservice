package config

import (
	"context"
	"github.com/ilyakaznacheev/cleanenv"
	"gitlab.com/adsfasdfdsf-group/key-value-service/pkg/logger"
)

type ROLE string

const (
	MASTER  ROLE = "master"
	REPLICA ROLE = "replica"
)

type Config struct {
	NodeID            string   `env:"NODE_ID"`
	PeerGRPCAddresses []string `env:"PEER_GRPC_ADDRESSES" env-separator:","`
	LeaderGRPCAddress string   `env:"LEADER_GRPC_ADDRESS"`
	PostgresDSN       string   `env:"POSTGRES_DSN" env-default:"postgres://root:123@localhost:5432/db?sslmode=disable"`
	RestPort          int      `env:"REST_PORT" env-default:"7070"`
	GrpcPort          int      `env:"GRPC_PORT" env-default:"4040"`
	ROLE              ROLE     `env:"ROLE" env-default:"replica"`
}

func New(ctx context.Context) *Config {
	cfg := &Config{}
	log := logger.GetLogger(ctx)
	err := cleanenv.ReadEnv(cfg)
	if err != nil {
		log.Error(ctx, err.Error())
	}
	return cfg
}
