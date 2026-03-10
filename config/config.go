package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct{
	DBUser string
	DBPassword string
	DBHost string
	DBPort string
	DBName string
	ServerPort string
	JWTSecret string
}

//LoadConfig reads the .env file and populate the Config struct

func LoadConfig() *Config{
	err := godotenv.Load()
	if err != nil{
		log.Println("Warning: No .env file found, reading from system environment variables")
	}

	return &Config{
		DBUser: os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_ROOT_PASSWORD"),
		DBHost: os.Getenv("DB_HOST"),
		DBPort: os.Getenv("DB_PORT"),
		DBName: os.Getenv("DB_NAME"),
		ServerPort: os.Getenv("SERVER_PORT"),
		JWTSecret : os.Getenv("jwtSecKey"),
	}
}