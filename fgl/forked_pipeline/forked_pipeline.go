package fgl_forked_pipeline

import (
	"time"

	"github.com/hkashwinkashyap/fastgoline/fgl/metadata"
	fgl_stage "github.com/hkashwinkashyap/fastgoline/fgl/stage"
)

// ForkedPipeline is struct of a forked pipeline
// It contains the id, input, stages, output, done, startedAt, and completedAt
// It is used to store the state of a forked pipeline
// It can be sorted once processed for timeseries processes
type ForkedPipeline[T any] struct {
	Id          string
	Input       metadata.ForkedInputMetadata[T]
	Stages      []fgl_stage.Stage[T]
	Output      chan metadata.ForkedOutputMetadata[T]
	Done        chan struct{}
	StartedAt   time.Time
	CompletedAt time.Time
}

// NewForkedPipeline creates a new forked pipeline instance
func NewForkedPipeline[T any](id string, input metadata.ForkedInputMetadata[T], stages []fgl_stage.Stage[T]) *ForkedPipeline[T] {
	return &ForkedPipeline[T]{
		Id:        id,
		Input:     input,
		Stages:    stages,
		Output:    make(chan metadata.ForkedOutputMetadata[T], 1),
		Done:      make(chan struct{}),
		StartedAt: time.Now(),
	}
}

// GetId returns the id of the forked pipeline
func (forkedPipeline *ForkedPipeline[T]) GetID() string {
	return forkedPipeline.Id
}

// GetInputMetadata returns the input metadata of the forked pipeline
func (forkedPipeline *ForkedPipeline[T]) GetInputMetadata() metadata.ForkedInputMetadata[T] {
	return forkedPipeline.Input
}

// GetStages returns the stages of the forked pipeline
func (forkedPipeline *ForkedPipeline[T]) GetStages() []fgl_stage.Stage[T] {
	return forkedPipeline.Stages
}

// GetOutput returns the output channel of the forked pipeline
func (forkedPipeline *ForkedPipeline[T]) GetOutput() <-chan metadata.ForkedOutputMetadata[T] {
	return forkedPipeline.Output
}

// IsDone returns true if the forked pipeline is done
func (forkedPipeline *ForkedPipeline[T]) IsDone() bool {
	select {
	case <-forkedPipeline.Done:
		return true
	default:
		return false
	}
}

// SetCompletedAt sets the completed at time of the forked pipeline
func (forkedPipeline *ForkedPipeline[T]) SetCompletedAt(timestamp time.Time) {
	forkedPipeline.CompletedAt = timestamp
}

// MarkDone marks the forked pipeline as done
func (forkedPipeline *ForkedPipeline[T]) MarkDone() {
	close(forkedPipeline.Done)
}
