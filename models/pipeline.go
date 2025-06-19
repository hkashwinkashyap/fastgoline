package models

import (
	"context"
	"fmt"
	"sync"
)

// Pipeline holds a sequence of stages to process data.
// Each stage runs concurrently, passing data along channels.
type Pipeline[T any] struct {
	id     int
	Stages []Stage[T]
}

// GetID returns the ID of the pipeline.
func (pipeline *Pipeline[T]) GetID() int {
	return pipeline.id
}

// NewPipeline creates an empty pipeline instance.
func NewPipeline[T any]() *Pipeline[T] {
	return &Pipeline[T]{}
}

// RunPipeline runs the pipeline to process data.
func (pipeline *Pipeline[T]) RunPipeline(ctx context.Context, in <-chan T, out chan<- T) error {
	var err error
	var wg sync.WaitGroup

	// Set the number of goroutines to the number of stages
	wg.Add(len(pipeline.Stages))

	currentIn := in

	for index, stage := range pipeline.Stages {
		currentOut := make(chan T)

		isFinal := index == len(pipeline.Stages)-1

		if isFinal {
			// Final stage
			// Pipe the final output to the provided output channel of the pipeline
			err = processStageTransform(ctx, stage, currentIn, out, &wg)
			if err != nil {
				ctx.Done()
				return err
			}
		} else {
			// Intermediate stage
			err = processStageTransform(ctx, stage, currentIn, currentOut, &wg)
			if err != nil {
				ctx.Done()
				return err
			}
		}

		// Current output channel becomes the next input channel
		currentIn = currentOut
	}

	// Wait for all stages to finish
	wg.Wait()

	return nil
}

// processStageTransform goroutine
func processStageTransform[T any](ctx context.Context, stage Stage[T], in <-chan T, out chan<- T, wg *sync.WaitGroup) error {
	go func() {
		defer wg.Done()

		err := stage.TransformFunction(ctx, in, out)
		close(out)
		if err != nil {
			fmt.Printf("ERR: %s\n", err.Error())
			return
		}
	}()

	return nil
}

// AddStage appends a processing stage to the pipeline.
func (pipeline *Pipeline[T]) AddStage(stage Stage[T]) {
	pipeline.Stages = append(pipeline.Stages, stage)
}

// RemoveStage removes a processing stage from the pipeline.
func (pipeline *Pipeline[T]) RemoveStage(stage Stage[T]) {
	for index, stageItem := range pipeline.Stages {
		if stageItem.id == stage.id {
			pipeline.Stages = append(pipeline.Stages[:index], pipeline.Stages[index+1:]...)
			break
		}
	}
}
