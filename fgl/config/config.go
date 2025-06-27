package config

import (
	"os"
)

// Config represents the configuration for the pipeline.
// TODO - more to come
type Config struct {
	LogLevel string
}

// getEnv returns the value of the environment variable if it is set,
// otherwise returns the default value
func getEnv(key string, defaultValue string) string {
	// Check if the environment variable is set
	envValue, valid := os.LookupEnv(key)
	if valid {
		// If the environment variable is set, return its value
		return envValue
	}

	// If the environment variable is not set, return the default value
	return defaultValue
}

// InitialiseConfig returns the configuration for the pipeline
// TODO - more to come
func InitialiseConfig() *Config {
	return &Config{
		LogLevel: getEnv("FGL_LOG_LEVEL", LogLevelDebug),
	}
}
