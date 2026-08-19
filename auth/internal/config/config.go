package config

import (
	"auth/pkg/logger"
	"context"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {

	//TODO db config
	//TODO grpc config or kafka config
	smth string
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