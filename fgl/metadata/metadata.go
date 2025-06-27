package metadata

import (
	"time"
)

// InputMetadata is struct of input metadata of a pipeline
// It contains the id, input value and input time
type InputMetadata[T any] struct {
	InitialInput *InitialInput[T]
	Value        T
	InputTime    time.Time
}

// InitialInput is struct of input metadata of a pipeline
// It contains the id, input value
type InitialInput[T any] struct {
	Id        string
	Value     T
	InputTime time.Time
}

// NewInputMetadata creates a new InputMetadata instance
// It takes in the input value as a parameter
func NewInputMetadata[T any](value T) InputMetadata[T] {
	inputTime := time.Now().UTC()

	return InputMetadata[T]{
		Value:     value,
		InputTime: inputTime,
	}
}

// OutputMetadata is the output struct of a pipeline
// It contains the initial input metadata, output value, output time, duration and error
// Duration is in nanoseconds
type OutputMetadata[T any] struct {
	InitialInput InitialInput[T]
	OutputValue  T
	OutputTime   time.Time
	Duration     int64
	Err          error
}
