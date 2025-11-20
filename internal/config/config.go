package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	HTTP struct {
		Port         int           `env:"HTTP_PORT" default:"8080"`
		ReadTimeout  time.Duration `env:"HTTP_READ_TIMEOUT" default:"5s"`
		WriteTimeout time.Duration `env:"HTTP_WRITE_TIMEOUT" default:"10s"`
	}
	DbUrl string `env:"DB_URL" env-required:"true"`
}

func MustLoad() *Config {
	_ = godotenv.Load()

	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("cannot read config: %v", err)
	}
	return &cfg
}
