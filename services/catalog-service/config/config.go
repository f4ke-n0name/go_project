package config

import (
	"fmt"
	"os"
)

type Config struct {
	DB   DBConfig
	GRPC GRPCConfig
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type GRPCConfig struct {
	Port string
}

func MustLoad() *Config {
	cfg := &Config{
		DB: DBConfig{
			Host:     mustGetEnv("DB_HOST"),
			Port:     mustGetEnv("DB_PORT"),
			User:     mustGetEnv("DB_USER"),
			Password: mustGetEnv("DB_PASSWORD"),
			Name:     mustGetEnv("DB_NAME"),
		},
		GRPC: GRPCConfig{
			Port: mustGetEnv("CATALOG_SERVICE_PORT"),
		},
	}

	return cfg
}

func (d *DBConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", d.User, d.Password, d.Host, d.Port, d.Name)
}

func mustGetEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic(fmt.Sprintf("environment variable %s must be set", key))
	}
	return val
}
