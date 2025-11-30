package config

import (
    "log"
    "os"

    "github.com/joho/godotenv"
)

type Config struct {
    AppPort   string
    PgHost    string
    PgPort    string
    PgUser    string
    PgPass    string
    PgDB      string
    MongoURI  string
    MongoDB   string
    JWTSecret string
}

func Load() *Config {
    _ = godotenv.Load()

    c := &Config{
        AppPort:   getEnv("APP_PORT", "8080"),
        PgHost:    getEnv("PG_HOST", "localhost"),
        PgPort:    getEnv("PG_PORT", "5432"),
        PgUser:    getEnv("PG_USER", "postgres"),
        PgPass:    getEnv("PG_PASSWORD", ""),
        PgDB:      getEnv("PG_DB", "prestasi_db"),
        MongoURI:  getEnv("MONGO_URI", "mongodb://localhost:27017"),
        MongoDB:   getEnv("MONGO_DB", "prestasi"),
        JWTSecret: getEnv("JWT_SECRET", "secret"),
    }

    log.Printf("Config loaded")
    return c
}

func getEnv(key, fallback string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return fallback
}
