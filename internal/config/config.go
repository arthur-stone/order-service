package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port            string
	Host            string
	Env             string
	LogLevel        string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

func Load() *Config {
	return &Config{
		Port:            getEnv("PORT", "8080"),
		Host:            getEnv("HOST", "0.0.0.0"),
		Env:             getEnv("ENV", "development"),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
		ReadTimeout:     getEnvDuration("READ_TIMEOUT_SEC", 10*time.Second),
		WriteTimeout:    getEnvDuration("WRITE_TIMEOUT_SEC", 10*time.Second),
		ShutdownTimeout: getEnvDuration("SHUTDOWN_TIMEOUT_SEC", 15*time.Second),
	}
}

func (c *Config) Addr() string {
	return c.Host + ":" + c.Port
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	valStr := os.Getenv(key)
	if valStr == "" {
		return defaultVal
	}
	sec, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultVal
	}
	return time.Duration(sec) * time.Second
}
