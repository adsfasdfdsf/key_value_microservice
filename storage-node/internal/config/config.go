package config

import (
	"context"
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
)

type ROLE string

const (
	MASTER ROLE = "master"
	REPLICA ROLE = "replica"
)

type Config struct{
	RestPort 	int 	`env:"REST_PORT" envDefault:"7070"`
	ROLE 		ROLE 	`env:"ROLE" envDefault:"replica"`	
}

func New(ctx context.Context) *Config{
	cfg := Config{}

	err := cleanenv.ReadEnv(cfg)
	if err != nil {
		fmt.Errorf("%v", err)
	}
	return &cfg
}

