package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hkashwinkashyap/fastgoline/fgl/metadata"
	fgl_pipeline "github.com/hkashwinkashyap/fastgoline/fgl/pipeline"
	fgl_stage "github.com/hkashwinkashyap/fastgoline/fgl/stage"
)

func main() {
	// Channels for pipeline 1
	in := make(chan metadata.InputMetadata[float64])
	out := make(chan metadata.OutputMetadata[float64], 3)

	pipeline1 := fgl_pipeline.NewPipeline[float64](in, out, false)

	// Channels for pipeline 2
	in2 := make(chan metadata.InputMetadata[float64])
	out2 := make(chan metadata.OutputMetadata[float64], 3)

	pipeline2 := fgl_pipeline.NewPipeline[float64](in2, out2, false)

	// Define reusable stage using StageTransformFunction
	// This is used when you want full control over input and output channels,
	// such as for aggregations (sum, average, filtering, etc.)
	total := fgl_stage.StageTransformFunction[float64](func(ctx context.Context, in <-chan metadata.InputMetadata[float64], out chan<- metadata.OutputMetadata[float64]) error {
		var total float64
		var output metadata.OutputMetadata[float64]

		for value := range in {
			output.InputID = value.Id
			output.InputTime = value.InputTime
			total += value.Value
		}

		output.Value = total

		out <- output
		return nil
	})

	// For simpler transformations (like activation functions or single-value transforms),
	// you can use NewStageFunction which abstracts channel handling.
	// These are suitable for stateless functions like scaling, normalization, etc.
	multiplyBy2 := fgl_stage.NewStageFunction[float64](func(value float64) float64 {
		return value * 2
	})

	percentage := fgl_stage.NewStageFunction[float64](func(value float64) float64 {
		return value / 100
	})

	// Add stages to each pipeline
	pipeline1.AddStage(multiplyBy2)
	pipeline1.AddStage(fgl_stage.NewStage[float64](total))
	pipeline1.AddStage(percentage)

	pipeline2.AddStage(multiplyBy2)
	pipeline2.AddStage(fgl_stage.NewStage[float64](total))

	// Run pipelines concurrently
	job := fgl_pipeline.PipelineJob[float64]{}

	job.AddPipeline(pipeline1)
	job.AddPipeline(pipeline2)

	startTime := time.Now().UTC()
	fmt.Printf("Starting pipelines at %s\n", startTime)

	// Launch all pipelines
	job.RunPipelinesInParallel(context.Background())

	// Pass in the input to the channels
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		intialValue := 10.0

		// Send input values to pipeline1
		for i := 0; i < 10000000; i++ {
			in <- metadata.NewInputMetadata(intialValue * 10.0)
			intialValue += 1.0
		}

		close(in)
	}()

	go func() {
		intialValue := 10.0

		// Send input values to pipeline2
		for i := 0; i < 1000000; i++ {
			in2 <- metadata.NewInputMetadata(intialValue / 10.0)
			intialValue -= 10.0
		}

		close(in2)
	}()

	go func() {
		for result := range out {
			fmt.Printf("Final result (pipeline1): %+v\n", result)
		}

		wg.Done()
	}()

	go func() {
		for result := range out2 {
			fmt.Printf("Final result (pipeline2): %+v\n", result)
		}

		wg.Done()
	}()

	wg.Wait()
}
