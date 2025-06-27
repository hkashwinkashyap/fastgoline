package fgl_stage

import (
	"context"
	"time"

	fgl_metadata "github.com/hkashwinkashyap/fastgoline/fgl/metadata"
	fgl_util "github.com/hkashwinkashyap/fastgoline/util"
)

// Stage represents a single stage in the pipeline.
type Stage[T any] struct {
	Id                string
	TransformFunction StageTransformFunction[T]
}

// StageTransformFunction represents a single processing stage transform function in the pipeline.
// It reads from the input channel, processes data, and sends results to the output channel.
// It returns an error if any stage fails.
type StageTransformFunction[T any] func(ctx context.Context, in chan fgl_metadata.InputMetadata[T], out chan fgl_metadata.OutputMetadata[T]) error

// NewStage creates a new Stage instance.
// Returns a Stage instance with a unique generated ID and the provided StageTransformFunction.
func NewStage[T any](stageTransformFunction StageTransformFunction[T]) Stage[T] {
	// Generate a unique id
	id := fgl_util.GenerateUUID()

	return Stage[T]{Id: id, TransformFunction: stageTransformFunction}
}

// GetID returns the ID of the stage.
func (stage *Stage[T]) GetID() string {
	return stage.Id
}

// NewStageFunction creates a new Stage instance.
// It takes in the transformation function as a parameter.
// Returns a Stage instance with a unique generated ID and the provided StageTransformFunction.
func NewStageFunction[T any](fn func(T) T) Stage[T] {
	transformFunction := func(ctx context.Context, in chan fgl_metadata.InputMetadata[T], out chan fgl_metadata.OutputMetadata[T]) error {
		for value := range in {
			// Format the output metadata and send it to the output channel
			out <- fgl_metadata.OutputMetadata[T]{
				InitialInput: fgl_metadata.InitialInput[T]{
					Id:        value.InitialInput.Id,
					Value:     value.InitialInput.Value,
					InputTime: value.InputTime,
				},
				OutputValue: fn(value.Value),
				OutputTime:  time.Now().UTC(),
				Duration:    time.Since(value.InputTime).Abs().Nanoseconds(),
				Err:         nil,
			}
		}

		return nil
	}

	// Return a new stage instance with the provided transformation function
	return NewStage(transformFunction)
}
