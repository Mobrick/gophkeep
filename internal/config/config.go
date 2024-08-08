package config

import (
	"flag"
	"os"
)

type Config struct {
	FlagRunAddr              string
	FlagLogLevel             string
	FlagDBConnectionAddress  string
}

func MakeConfig() *Config {
	config := &Config{}

	flag.StringVar(&config.FlagRunAddr, "a", ":8080", "address to run server")
	flag.StringVar(&config.FlagLogLevel, "l", "info", "log level")
	flag.StringVar(&config.FlagDBConnectionAddress, "d", "host=localhost port=5432 user=postgres password=vvv dbname=gophkeep sslmode=disable", "database connection address")

	flag.Parse()

	if envRunAddr := os.Getenv("RUN_ADDRESS"); envRunAddr != "" {
		config.FlagRunAddr = envRunAddr
	}

	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		config.FlagLogLevel = envLogLevel
	}

	if envDBConnectionAddress := os.Getenv("DATABASE_URI"); envDBConnectionAddress != "" {
		config.FlagDBConnectionAddress = envDBConnectionAddress
	}
	return config
}