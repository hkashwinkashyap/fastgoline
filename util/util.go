package util

import (
	"github.com/google/uuid"
)

// GenerateUUID generates and returns a unique ID
// This is used to generate unique IDs for stages and pipelines
// It uses the ulid library to generate a unique ID which is a combination of timestamp and entropy
func GenerateUUID() string {
	// Generate UUID
	id := uuid.New().String()

	return id
}
