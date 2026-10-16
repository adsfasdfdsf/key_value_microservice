package config

import (
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	AuthPostgreConfig
	StorageURL string `env:"STORAGE_URL" env-default:"http://localhost"`
}

func ReadFromFile() (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
