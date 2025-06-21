package fgl

import (
	"context"
	"fmt"
	"sync"
)

// PipelineJob represents a job to run a pipeline.
// It contains an arrray of pipelines to run in parallel.
type PipelineJob[T any] struct {
	Pipelines []Pipeline[T]
}

// NewPipelineJob creates a new PipelineJob instance.
func NewPipelineJob[T any]() *PipelineJob[T] {
	return &PipelineJob[T]{}
}

// AddPipeline adds a pipeline to the job.
func (pipelineJob *PipelineJob[T]) AddPipeline(pipeline *Pipeline[T]) {
	pipelineJob.Pipelines = append(pipelineJob.Pipelines, *pipeline)
}

// RunPipelinesInParallel runs all of pipelines in the job in parallel.
// It returns an error if any of the pipelines fails.
func (pipelineJob *PipelineJob[T]) RunPipelinesInParallel(ctx context.Context) {
	defer func() {
		ctx.Done()
	}()

	// Run all pipelines in parallel
	go func() {
		var wg sync.WaitGroup
		wg.Add(len(pipelineJob.Pipelines))

		fmt.Println("Running", len(pipelineJob.Pipelines), "pipelines in parallel...")

		// Create a channel to collect errors
		errChan := make(chan error, len(pipelineJob.Pipelines))

		// Loop through all pipelines
		for _, pipeline := range pipelineJob.Pipelines {
			// Kick off a goroutine for each pipeline
			go func(pipeline *Pipeline[T]) {
				defer wg.Done()

				// Run the pipeline
				err := pipeline.RunPipeline(ctx)
				if err != nil {
					// Send the error to the error channel
					errChan <- err
				}
			}(&pipeline)
		}

		wg.Wait()
		close(errChan)

		// Collect and return first error
		if len(errChan) > 0 {
			panic(<-errChan)
		}

		// Return success
		ctx.Done()
	}()
}
