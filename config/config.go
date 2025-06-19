package config

import "os"

type Config struct {
	MongoURI  string
	JwtSecret string
	DbName    string
	Port      string
}

func Load() *Config {
	return &Config{
		MongoURI:  os.Getenv("MONGO_URI"),
		JwtSecret: os.Getenv("JWT_SECRET"),
		DbName:    os.Getenv("MONGO_DBNAME"),
		Port:      getEnv("PORT", "8081"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
