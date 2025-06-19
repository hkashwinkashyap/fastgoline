package models

import (
	"context"
	"math/rand"
	"time"

	"github.com/oklog/ulid"
)

// Stage represents a single stage in the pipeline.
type Stage[T any] struct {
    id    string 
    TransformFunction StageTransformFunction[T] 
}

// StageTransformFunction represents a single processing stage transform function in the pipeline.
// It reads from the input channel, processes data, and sends results to the output channel.
type StageTransformFunction[T any] func(ctx context.Context, in <-chan T, out chan<- T) error

// NewStage creates a new Stage instance.
// Returns a Stage instance with a unique generated ID and the provided StageTransformFunction.
func NewStage[T any](stageTransformFunction StageTransformFunction[T]) Stage[T] {
    // Generate Unique ID
    t := time.Unix(1000000, 0)
    entropy := ulid.Monotonic(rand.New(rand.NewSource(t.UnixNano())), 0)
    id := ulid.MustNew(ulid.Timestamp(t), entropy).String()

    return Stage[T]{id: id, TransformFunction: stageTransformFunction}
}

// GetID returns the ID of the stage.
func (stage *Stage[T]) GetID() string {
    return stage.id
}