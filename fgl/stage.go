package fgl

import (
	"context"
	"fastgoline/util"
)

// Stage represents a single stage in the pipeline.
type Stage[T any] struct {
	id                string
	TransformFunction StageTransformFunction[T]
}

// StageTransformFunction represents a single processing stage transform function in the pipeline.
// It reads from the input channel, processes data, and sends results to the output channel.
// It returns an error if any stage fails.
type StageTransformFunction[T any] func(ctx context.Context, in <-chan T, out chan<- T) error

// NewStage creates a new Stage instance.
// Returns a Stage instance with a unique generated ID and the provided StageTransformFunction.
func NewStage[T any](stageTransformFunction StageTransformFunction[T]) Stage[T] {
	// Generate a unique id
	id := util.GenerateUUID()

	return Stage[T]{id: id, TransformFunction: stageTransformFunction}
}

// GetID returns the ID of the stage.
func (stage *Stage[T]) GetID() string {
	return stage.id
}
