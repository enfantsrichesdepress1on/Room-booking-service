package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port              string
	DBHost            string
	DBPort            string
	DBName            string
	DBUser            string
	DBPassword        string
	DBSSLMode         string
	JWTSecret         string
	SlotWindowDays    int
	ConferenceBaseURL string
}

func Load() Config {
	return Config{
		Port:              getEnv("APP_PORT", "8080"),
		DBHost:            getEnv("DB_HOST", "localhost"),
		DBPort:            getEnv("DB_PORT", "5432"),
		DBName:            getEnv("DB_NAME", "room_booking"),
		DBUser:            getEnv("DB_USER", "postgres"),
		DBPassword:        getEnv("DB_PASSWORD", "postgres"),
		DBSSLMode:         getEnv("DB_SSLMODE", "disable"),
		JWTSecret:         getEnv("JWT_SECRET", "supersecret"),
		ConferenceBaseURL: getEnv("CONFERENCE_BASE_URL", "https://conference.mock.local/meeting"),
		SlotWindowDays:    14,
	}
}

func (c Config) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
