package config

import (
	"fmt"
	"os"
)

type Config struct {
	HTTP     HTTPConfig
	Services ServicesConfig
}

type HTTPConfig struct {
	Port string
}

type ServicesConfig struct {
	AuthAddr      string
	CatalogAddr   string
	OrderAddr     string
	InventoryAddr string
}

func MustLoad() *Config {
	return &Config{
		HTTP: HTTPConfig{
			Port: mustGetEnv("GATEWAY_PORT"),
		},
		Services: ServicesConfig{
			AuthAddr:      mustGetEnv("AUTH_SERVICE_ADDR"),
			CatalogAddr:   mustGetEnv("CATALOG_SERVICE_ADDR"),
			OrderAddr:     mustGetEnv("ORDER_SERVICE_ADDR"),
			InventoryAddr: mustGetEnv("INVENTORY_SERVICE_ADDR"),
		},
	}
}

func mustGetEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic(fmt.Sprintf("required env variable %q is not set", key))
	}
	return val
}
