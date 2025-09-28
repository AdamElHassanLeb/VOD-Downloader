package Env

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
)

func init() {
	// Load .Env only once during package initialization
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: Error loading .Env file: %s", err)
	}
}

func GetString(key, fallback string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	return val
}

func GetInt(key string, fallback int) int {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	valAsInt, err := strconv.Atoi(val)
	if err != nil {
		return fallback
	}
	return valAsInt
}
