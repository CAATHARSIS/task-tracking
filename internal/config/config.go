package congig

import (
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	// Общие настройки
	AppEnv  string `envcoding:"APP_ENV" default:"development"`
	AppPort string `envcoding:"APP_PORT" default:"8080"`

	// Настройки базы данных
	DBHost     string `envcoding:"DB_HOST" default:"localhost"`
	DBPort     string `envcoding:"DP_PORT" default:"5432"`
	DBUser     string `envcoding:"DB_USER" default:"postgres"`
	DBPassword string `envcoding:"DB_PASSWORD" default:"postgres"`
	DBName     string `envcoding:"DB_NAME" default:"task-tracking"`
	DBSSLMode  string `envcoding:"DBSSLMODE" default:"disable"`

	// Настройки Redis
	RedisHost     string `envcoding:"REDIS_HOST" default:"localhost"`
	RedisPort     string `envcoding:"REDIS_PORT" default:"63879"`
	RedisPassword string `envcoding:"REDIS_PASSWORD" default:""`
	RedisDB       string `envcoding:"REDIS_DB" default:"0"`

	// Настройки JWT
	JWTSecret     string        `envcoding:"JWT_SECRET" required:"true"`
	JWTExpiration time.Duration `envcoding:"JWT_EXPIRATION" default:"24h"`

	// Настройки миграций
	MigrationsPath string `envcoding:"MIGRATIONS_PATH" default:"file://migrations"`
}

func Load() (*Config, error) {
	if env := os.Getenv("APP_ENV"); env == "" || env == "development" {
		_ = godotenv.Load()
	}

	var cfg Config
	err := envconfig.Process("", &cfg)
	return &cfg, err
}
