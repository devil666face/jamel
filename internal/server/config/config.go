package config

import (
	"fmt"
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	GrpcIP       string `env:"GRPC_IP" env-default:"0.0.0.0"`
	GrpcPort     uint   `env:"GRPC_PORT" env-default:"8443"`
	SqliteDB     string `env:"SQLITE_DB" env-default:"db/db.sqlite3"`
	AuthUsername string `env:"AUTH_USERNAME" env-default:"user"`
	AuthPassword string `env:"AUTH_PASSWORD" env-default:"Qwerty123!"`
	LogFile      string `env:"LOG_FILE" env-default:"jamel-server.log"`
	GrpcConnect  string
}

func Must() *Config {
	_config := Config{}
	if err := cleanenv.ReadConfig(".env", &_config); err != nil {
		if err := cleanenv.ReadEnv(&_config); err != nil {
			log.Fatalf("env variable not found: %s", err)
		}
	}
	_config.GrpcConnect = fmt.Sprintf("%s:%d", _config.GrpcIP, _config.GrpcPort)
	return &_config
}
