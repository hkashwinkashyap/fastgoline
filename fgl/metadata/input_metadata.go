package metadata

import "time"

// ForkedInputMetadata is struct of input metadata of a forked pipeline
// It contains the id, value, and created at
type ForkedInputMetadata[T any] struct {
	Id        string
	Value     T
	CreatedAt time.Time
}

// ForkedOutputMetadata is the output struct of a forked pipeline
// It contains the input ID, output, duration, and error
type ForkedOutputMetadata[T any] struct {
	InputID     string
	Output      T
	CompletedAt time.Time
	Duration    time.Duration
	Err         error
}
