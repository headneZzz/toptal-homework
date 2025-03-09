package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type DatabaseConfig struct {
	Host         string
	Port         string
	User         string
	Password     string
	Name         string
	MaxOpenConns int
	MaxIdleConns int
	MaxLifetime  time.Duration
	SSLMode      string
}

type ServerConfig struct {
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

type MetricsConfig struct {
	Enabled bool
	Port    string
}

type SecurityConfig struct {
	JWTSecret          string
	JWTExpirationHours int
	BcryptCost         int
}

type CartConfig struct {
	CleanupInterval time.Duration
	ExpiryTime      time.Duration
}

type LogConfig struct {
	Level string
	JSON  bool
}

type Config struct {
	Environment string
	DB          DatabaseConfig
	Server      ServerConfig
	Metrics     MetricsConfig
	Security    SecurityConfig
	Cart        CartConfig
	Log         LogConfig
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		slog.Warn("Failed to load .env file, using environment variables", "error", err)
	}

	return &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		DB: DatabaseConfig{
			Host:         getEnv("DB_HOST", "localhost"),
			Port:         getEnv("DB_PORT", "5432"),
			User:         getEnv("DB_USER", "postgres"),
			Password:     getEnv("DB_PASSWORD", ""),
			Name:         getEnv("DB_NAME", "bookshop"),
			MaxOpenConns: getEnvAsInt("DB_MAX_OPEN_CONNS", 10),
			MaxIdleConns: getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
			MaxLifetime:  getEnvAsDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
			SSLMode:      getEnv("DB_SSL_MODE", "disable"),
		},
		Server: ServerConfig{
			Port:            getEnv("SERVER_PORT", "8080"),
			ReadTimeout:     getEnvAsDuration("SERVER_READ_TIMEOUT", 10*time.Second),
			WriteTimeout:    getEnvAsDuration("SERVER_WRITE_TIMEOUT", 10*time.Second),
			ShutdownTimeout: getEnvAsDuration("SERVER_SHUTDOWN_TIMEOUT", 30*time.Second),
		},
		Metrics: MetricsConfig{
			Enabled: getEnvAsBool("METRICS_ENABLED", true),
			Port:    getEnv("METRICS_PORT", "2112"),
		},
		Security: SecurityConfig{
			JWTSecret:          getEnv("JWT_SECRET", "your_secret_key"),
			JWTExpirationHours: getEnvAsInt("JWT_EXPIRATION_HOURS", 24),
			BcryptCost:         getEnvAsInt("BCRYPT_COST", 10),
		},
		Cart: CartConfig{
			CleanupInterval: getEnvAsDuration("CART_CLEANUP_INTERVAL", 5*time.Minute),
			ExpiryTime:      getEnvAsDuration("CART_EXPIRY_TIME", 30*time.Minute),
		},
		Log: LogConfig{
			Level: getEnv("LOG_LEVEL", "info"),
			JSON:  getEnvAsBool("LOG_JSON", true),
		},
	}, nil
}

func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.Name, c.SSLMode)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
