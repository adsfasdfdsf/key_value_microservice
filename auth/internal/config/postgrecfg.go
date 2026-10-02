package config

type AuthPostgreConfig struct {
	UserName string `env:"POSTGRES_USER" env-default:"root"`
	Password string `env:"POSTGRES_PASSWORD" env-default:"123"`
	Host     string `env:"POSTGRES_HOST" env-default:"auth_db"`
	Port     string `env:"POSTGRES_PORT" env-default:"5432"`
	DbName   string `env:"POSTGRES_DB" env-default:"db"`
}