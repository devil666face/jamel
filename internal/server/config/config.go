package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	GrpcConnect  string `env:"GRPC_CONNECT" env-default:"127.0.0.1:8443"`
	SqliteDB     string `env:"SQLITE_DB" env-default:"db/db.sqlite3"`
	AuthUsername string `env:"AUTH_USERNAME" env-default:"user"`
	AuthPassword string `env:"AUTH_PASSWORD" env-default:"Qwerty123!"`
	LogFile      string `env:"LOG_FILE" env-default:"jamel-server.log"`
	S3Conntect   string `env:"S3_CONNECT" env-default:"127.0.0.1:9000"`
	S3Username   string `env:"S3_USERNAME" env-default:"minio"`
	S3Password   string `env:"S3_PASSWORD" env-default:"password"`
	S3Bucket     string `env:"S3_BUCKET" env-default:"jamel"`
	RMQConnect   string `env:"RMQ_CONNECT" env-default:"127.0.0.1:5672"`
	RMQUsername  string `env:"RMQ_USERNAME" env-default:"rabbitmq"`
	RMQPassword  string `env:"RMQ_PASSWORD" env-default:"password"`
}

func Must() *Config {
	_config := Config{}
	if err := cleanenv.ReadConfig(".env", &_config); err != nil {
		if err := cleanenv.ReadEnv(&_config); err != nil {
			log.Fatalf("env variable not found: %s", err)
		}
	}
	return &_config
}
