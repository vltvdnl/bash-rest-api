package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

// default environment variables
const (
	default_host     = "localhost"
	default_port     = "5432"
	default_user     = "postgres"
	default_password = "postgres"
	default_name     = "postgres"
	default_sslmode  = "disable"
)

// init считывает данные из .env
func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, will be used default values")
	}
}

// PostgreConfig хранит данные, необходимые для поключения к Postgres
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
		Host:     getEnv("DB_HOST", default_host),
		Port:     getEnv("DB_PORT", default_port),
		User:     getEnv("DB_USER", default_user),
		Password: getEnv("DB_PASSWORD", default_password),
		DBname:   getEnv("DB_NAME", default_name),
		SSLmode:  getEnv("SSLmode", default_sslmode),
	}
}

// getEnv считывает данные из переменных окружения
func getEnv(key string, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue // если нет переменной с именем key
}

// String возвращает строку (dataSourceName), необходимую для подключение к postgres
func (c PostgreConfig) String() string {
	return fmt.Sprintf("host = %s port=%s user=%s password = %s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBname, c.SSLmode)
}
