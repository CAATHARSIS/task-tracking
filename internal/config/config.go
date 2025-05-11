package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	// Общие настройки
	AppEnv  string `envconfig:"APP_ENV" default:"development"`
	AppPort string `envconfig:"APP_PORT" default:"8080"`

	// Настройки базы данных
	DBHost     string `envconfig:"DB_HOST" default:"localhost"`
	DBPort     string `envconfig:"DB_PORT" default:"5432"`
	DBUser     string `envconfig:"DB_USER" default:"postgres"`
	DBPassword string `envconfig:"DB_PASSWORD" default:"postgres"`
	DBName     string `envconfig:"DB_NAME" default:"task-tracking"`
	DBSSLMode  string `envconfig:"DBSSLMODE" default:"disable"`

	// Настройки Redis
	RedisHost     string `envcong:"REDIS_HOST" default:"localhost"`
	RedisPort     string `envconfig:"REDIS_PORT" default:"6379"`
	RedisPassword string `envconfig:"REDIS_PASSWORD" default:""`
	RedisDB       string `envconfig:"REDIS_DB" default:"0"`

	// Настройки JWT
	JWTSecret     string        `envconfig:"JWT_SECRET" required:"true"`
	JWTExpiration time.Duration `envconfig:"JWT_EXPIRATION" default:"24h"`

	// Настройки миграций
	MigrationsPath string `envconfig:"MIGRATIONS_PATH" default:"file://migrations"`
}

func Load() (*Config, error) {
	if env := os.Getenv("APP_ENV"); env == "" || env == "development" {
		_ = godotenv.Load()
	}

	var cfg Config
	err := envconfig.Process("", &cfg)
	return &cfg, err
}
