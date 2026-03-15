package config

import "os"

type Config struct {
	Port               string
	DBURL              string
	KafkaBrokers       string
	MatchingEngineAddr string
}

func Load() *Config {
	return &Config{
		Port:               getEnv("PORT", "8081"),
		DBURL:              getEnv("DB_URL", "postgres://mantis:mantis@localhost:5432/mantis_market?sslmode=disable"),
		KafkaBrokers:       getEnv("KAFKA_BROKERS", "localhost:9092"),
		MatchingEngineAddr: getEnv("MATCHING_ENGINE_ADDR", "localhost:50051"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
