package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
	"time"
)

type Config struct {
	ConfigPath string `env:"CONFIG_PATH" env-default:"config/config.yaml"`
	Database   Database
	HttpServer HttpServer `yaml:"HttpServer"`
}

type Database struct {
	DbHost string `env:"DB_HOST" env-required:"true"`
	DbPort int    `env:"DB_PORT" env-required:"true"`
	DbUser string `env:"DB_USER" env-required:"true"`
	DbPass string `env:"DB_PASS" env-required:"true"`
	DbName string `env:"DB_NAME" env-required:"true"`
}

type HttpServer struct {
	Address      int           `yaml:"Address" env-default:"8080"`
	IdleTimeout  time.Duration `yaml:"idle_timeout" yaml-default:"60s"`
	ReadTimeout  time.Duration `yaml:"read_timeout" yaml-default:"10s"`
	WriteTimeout time.Duration `yaml:"write_timeout" yaml-default:"10s"`
}

func Load() *Config {
	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Printf("cannot read .env file: %s\n (fix: you need to put .env file in main dir)", err)
		os.Exit(1)
	}
	return &cfg
}
