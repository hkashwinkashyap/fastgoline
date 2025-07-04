package config

import (
	"os"
	"runtime"
	"strconv"
)

// Config represents the configuration for the pipeline.
type Config struct {
	LogLevel    string
	MaxWorkers  int
	MaxMemoryMB uint64
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
// Default LogLevel = LogLevelDebug
// Default MaxWorkers = number of CPUs * 2
// Default MaxMemoryMB = 1024
func InitialiseConfig() *Config {
	maxWorkers, err := strconv.Atoi(getEnv("FGL_MAX_WORKERS", strconv.Itoa(runtime.NumCPU()*2)))
	if err != nil {
		maxWorkers = runtime.NumCPU() * 2
	}

	maxMemoryMB, err := strconv.ParseUint(getEnv("FGL_MAX_MEMORY_MB", "1024"), 10, 64)
	if err != nil {
		maxMemoryMB = 1024
	}

	return &Config{
		LogLevel:    getEnv("FGL_LOG_LEVEL", string(LogLevelDebug)),
		MaxWorkers:  maxWorkers,
		MaxMemoryMB: maxMemoryMB,
	}
}
