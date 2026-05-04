package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port       string
	BasePath   string
	JWTSecret  string
	UploadDir  string
	MySQLDSN   string
	ProvAPIURL string
}

func Load() Config {
	_ = godotenv.Load()

	port := env("APP_PORT", "8000")
	return Config{
		Port:       port,
		BasePath:   env("APP_BASE_PATH", "/api/v1"),
		JWTSecret:  env("JWT_SECRET", "dev-secret-change-me"),
		UploadDir:  env("UPLOAD_DIR", "uploads"),
		MySQLDSN:   mysqlDSN(),
		ProvAPIURL: env("PROVCITY_API_URL", "https://www.emsifa.com/api-wilayah-indonesia/api"),
	}
}

func mysqlDSN() string {
	if dsn := os.Getenv("MYSQL_DSN"); dsn != "" {
		return dsn
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=%s&loc=%s",
		env("DB_USER", "root"),
		env("DB_PASSWORD", ""),
		env("DB_HOST", "127.0.0.1"),
		env("DB_PORT", "3306"),
		env("DB_NAME", "rkmin"),
		env("DB_PARSE_TIME", "true"),
		env("DB_LOC", "Local"),
	)
}

func env(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
