package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Server   string `env:"SERVER" env-default:"127.0.0.1:8443"`
	Username string `env:"LOGIN" env-default:"admin"`
	Password string `env:"PASSWORD" env-default:"YwuPCnqUqvqz563#$%!@^"`
}

func Must() *Config {
	cfg := Config{}
	if err := cleanenv.ReadConfig(".env", &cfg); err != nil {
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			log.Fatalf("env variable not found: %v", err)
		}
	}
	return &cfg
}
