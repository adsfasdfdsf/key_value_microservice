package config

import (
	"context"
	"github.com/ilyakaznacheev/cleanenv"
	"gitlab.com/adsfasdfdsf-group/key-value-service/pkg/logger"
)

type ROLE string

const (
	MASTER ROLE = "master"
	REPLICA ROLE = "replica"
)

type Config struct{
	RestPort 	int 	`env:"REST_PORT" envDefault:"7070"`
	GrpcPort 	int 	`env:"GRPC_PORT" envDefault:"4040"`
	ROLE 		ROLE 	`env:"ROLE" envDefault:"replica"`	
}

func New(ctx context.Context) *Config{
	cfg := &Config{}
	log := logger.GetLogger(ctx)
	err := cleanenv.ReadEnv(cfg)
	if err != nil {
		log.Error(ctx, err.Error())
	}
	return cfg
}

