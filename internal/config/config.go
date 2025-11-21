package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	HTTP struct {
		Address      string        `env:"HTTP_ADDRESS" env-required:"true"`
		ReadTimeout  time.Duration `env:"HTTP_READ_TIMEOUT"`
		WriteTimeout time.Duration `env:"HTTP_WRITE_TIMEOUT"`
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
