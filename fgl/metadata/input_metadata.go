package metadata

import "time"

type InputMetadata[T any] struct {
	Id        string
	Value     T
	StartedAt time.Time
}
