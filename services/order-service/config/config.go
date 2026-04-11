package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	DB    DBConfig
	GRPC  GRPCConfig
	Kafka KafkaConfig
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

type KafkaConfig struct {
	Brokers []string
}

func MustLoad() *Config {
	return &Config{
		DB: DBConfig{
			Host:     mustGetEnv("DB_HOST"),
			Port:     mustGetEnv("DB_PORT"),
			User:     mustGetEnv("DB_USER"),
			Password: mustGetEnv("DB_PASSWORD"),
			Name:     mustGetEnv("DB_NAME"),
		},
		GRPC: GRPCConfig{
			Port: mustGetEnv("ORDER_SERVICE_PORT"),
		},
		Kafka: KafkaConfig{
			Brokers: strings.Split(mustGetEnv("KAFKA_BROKERS"), ","),
		},
	}
}

func (c *DBConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.User, c.Password, c.Host, c.Port, c.Name,
	)
}

func mustGetEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic(fmt.Sprintf("required env variable %q is not set", key))
	}
	return val
}
