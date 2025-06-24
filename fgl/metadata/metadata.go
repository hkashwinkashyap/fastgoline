package metadata

import (
	"time"

	"github.com/hkashwinkashyap/fastgoline/util"
)

// InputMetadata is struct of input metadata of a pipeline
// It contains the id, input value and input time
type InputMetadata[T any] struct {
	Id        string
	Value     T
	InputTime time.Time
}

// NewInputMetadata creates a new InputMetadata instance
// It takes in the input value as a parameter and generates a unique id
func NewInputMetadata[T any](value T) InputMetadata[T] {
	id := util.GenerateUUID()

	inputTime := time.Now().UTC()

	return InputMetadata[T]{
		Id:        id,
		Value:     value,
		InputTime: inputTime,
	}
}

// OutputMetadata is the output struct of a pipeline
// It contains the input ID, output value, input time, output time, duration and error
// Duration is in nanoseconds
// NOTE: OutputTime will be invalid if the last stage is an aggregation transformation
type OutputMetadata[T any] struct {
	InputID    string
	Value      T
	InputTime  time.Time
	OutputTime time.Time
	ForkedAt   *time.Time
	Duration   int64
	Err        error
}
