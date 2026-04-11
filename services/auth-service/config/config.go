package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	DB   DBConfig
	JWT  JWTConfig
	GRPC GRPCConfig
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}
type JWTConfig struct {
	Secret  string
	Access  time.Duration
	Refresh time.Duration
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
		JWT: JWTConfig{
			Secret:  mustGetEnv("JWT_SECRET"),
			Access:  mustGetDuration("JWT_ACCESS", 15*time.Minute),
			Refresh: mustGetDuration("JWT_REFRESH", 7*24*time.Hour),
		},
		GRPC: GRPCConfig{
			Port: mustGetEnv("AUTH_SERVICE_PORT"),
		},
	}

	return cfg
}

func mustGetEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic(fmt.Sprintf("environment variable %s must be set", key))
	}
	return val
}

func mustGetDuration(key string, defaultVal time.Duration) time.Duration {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	d, err := time.ParseDuration(val)
	if err != nil {
		panic(fmt.Sprintf("environment variable %s must be a valid duration", key))
	}
	return d
}

func (d *DBConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", d.User, d.Password, d.Host, d.Port, d.Name)
}
