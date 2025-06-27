package fgl_pipeline

import (
	"context"
	"fmt"
	"sync"

	fgl_config "github.com/hkashwinkashyap/fastgoline/fgl/config"
)

// PipelineJob represents a job to run a pipeline.
// It contains an arrray of pipelines to run in parallel.
type PipelineJob[T any] struct {
	Pipelines []Pipeline[T]
	Config    fgl_config.Config
}

// NewPipelineJob creates a new PipelineJob instance.
func NewPipelineJob[T any](config *fgl_config.Config) *PipelineJob[T] {
	if config == nil {
		config = fgl_config.InitialiseConfig()
	}

	return &PipelineJob[T]{
		Config: *config,
	}
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

		if pipelineJob.Config.LogLevel == "DEBUG" {
			fmt.Println("TRACE: Running", len(pipelineJob.Pipelines), "pipelines in parallel...")
		}

		// Loop through all pipelines
		for _, pipeline := range pipelineJob.Pipelines {
			// Kick off a goroutine for each pipeline
			go func(pipeline *Pipeline[T]) {
				defer wg.Done()

				// Run the pipeline
				pipeline.RunPipeline(ctx)
			}(&pipeline)
		}

		wg.Wait()

		// Return success
		ctx.Done()
	}()
}
