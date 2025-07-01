package fgl_stage

import (
	"context"
	"time"

	fgl_metadata "github.com/hkashwinkashyap/fastgoline/fgl/metadata"
	fgl_util "github.com/hkashwinkashyap/fastgoline/util"
)

// Stage represents a single stage in the pipeline.
type Stage[T any] struct {
	id                string
	TransformFunction StageTransformFunction[T]
}

// StageTransformFunction represents a single processing stage transform function in the pipeline.
// It reads from the input, processes data, and return the results.
// It returns an error if any stage fails.
type StageTransformFunction[T any] func(ctx context.Context, input fgl_metadata.InputMetadata[T]) fgl_metadata.OutputMetadata[T]

// NewStage creates a new Stage instance.
// Returns a Stage instance with a unique generated ID and the provided StageTransformFunction.
func NewStage[T any](stageTransformFunction StageTransformFunction[T]) Stage[T] {
	// Generate a unique id
	id := fgl_util.GenerateUUID()

	return Stage[T]{id: id, TransformFunction: stageTransformFunction}
}

// GetID returns the ID of the stage.
func (stage *Stage[T]) GetID() string {
	return stage.id
}

// NewStageFunction creates a new Stage instance.
// It takes in the transformation function as a parameter.
// Returns a Stage instance with a unique generated ID and the provided StageTransformFunction.
func NewStageFunction[T any](fn func(T) (T, error)) Stage[T] {
	transformFunction := func(ctx context.Context, input fgl_metadata.InputMetadata[T]) fgl_metadata.OutputMetadata[T] {
		outputValue, err := fn(input.Value)
		if err != nil {
			return fgl_metadata.OutputMetadata[T]{
				InitialInput: fgl_metadata.InitialInput[T]{
					Id:        input.InitialInput.Id,
					Value:     input.InitialInput.Value,
					InputTime: input.InputTime,
				},
				OutputValue: outputValue,
				OutputTime:  time.Now().UTC(),
				Duration:    time.Since(input.InputTime).Abs().Nanoseconds(),
				Err:         err,
			}
		}

		return fgl_metadata.OutputMetadata[T]{
			InitialInput: fgl_metadata.InitialInput[T]{
				Id:        input.InitialInput.Id,
				Value:     input.InitialInput.Value,
				InputTime: input.InputTime,
			},
			OutputValue: outputValue,
			OutputTime:  time.Now().UTC(),
			Duration:    time.Since(input.InputTime).Abs().Nanoseconds(),
			Err:         nil,
		}
	}

	// Return a new stage instance with the provided transformation function
	return NewStage(transformFunction)
}
