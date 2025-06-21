package util

import (
	"math/rand"
	"time"

	"github.com/oklog/ulid"
)

// GenerateUUID generates and returns a unique ID
// This is used to generate unique IDs for stages and pipelines
// It uses the ulid library to generate a unique ID which is a combination of timestamp and entropy
func GenerateUUID() string {
	// Generate Unique ID
	t := time.Unix(1000000, 0)
	entropy := ulid.Monotonic(rand.New(rand.NewSource(t.UnixNano())), 0)
	id := ulid.MustNew(ulid.Timestamp(t), entropy).String()

	return id
}
