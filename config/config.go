package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
}

type PostgreConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBname   string
	SSLmode  string
}

func New() *PostgreConfig {
	return &PostgreConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "postgres"),
		DBname:   getEnv("DB_NAME", "postgres"),
		SSLmode:  getEnv("SSLmode", "disable"),
	}
}

func getEnv(key string, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}

func (c PostgreConfig) String() string {
	return fmt.Sprintf("host = %s port=%s user=%s password = %s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBname, c.SSLmode)
}
