package main

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	fgl_stage "github.com/hkashwinkashyap/fastgoline/fgl/stage"

	fgl_pipeline "github.com/hkashwinkashyap/fastgoline/fgl/pipeline"
)

func main() {
	// Create a new pipeline
	in := make(chan float64)
	out := make(chan float64, 3)

	pipeline := fgl_pipeline.NewPipeline[float64](in, out)

	wg := sync.WaitGroup{}
	wg.Add(1)

	// Pipeline 2
	in2 := make(chan float64)
	out2 := make(chan float64, 3)

	pipeline2 := fgl_pipeline.NewPipeline[float64](in2, out2)

	wg2 := sync.WaitGroup{}
	wg2.Add(1)

	// Define stage functions
	// multiplyBy2 := fgl_stage.StageTransformFunction[float64](func(ctx context.Context, in <-chan float64, out chan<- float64) error {
	// 	for value := range in {
	// 		out <- value * 2
	// 	}

	// 	return nil
	// })

	// percentage := fgl.StageTransformFunction[float64](func(ctx context.Context, in <-chan float64, out chan<- float64) error {
	// 	for value := range in {
	// 		out <- value / 100
	// 	}

	// 	return nil
	// })

	total := fgl_stage.StageTransformFunction[float64](func(ctx context.Context, in <-chan float64, out chan<- float64) error {
		var total float64
		for value := range in {
			total += value
		}

		out <- math.Floor(total*100) / 100
		return nil
	})

	// // Add stages to the pipeline
	// pipeline.AddStage(fgl.NewStage[float64](multiplyBy2))
	// pipeline.AddStage(fgl.NewStage[float64](percentage))
	// pipeline.AddStage(fgl.NewStage[float64](total))

	// // Add stages to the pipeline2 (reused stages)
	// pipeline2.AddStage(fgl.NewStage[float64](multiplyBy2))
	// pipeline2.AddStage(fgl.NewStage[float64](total))

	multiplyBy2 := fgl_stage.NewStageFunction[float64](func(value float64) float64 {
		return value * 2
	})

	percentage := fgl_stage.NewStageFunction[float64](func(value float64) float64 {
		return value / 100
	})

	// Add stages to the pipeline
	pipeline.AddStage(multiplyBy2)
	pipeline.AddStage(fgl_stage.NewStage[float64](total))
	pipeline.AddStage(percentage)

	// Add stages to the pipeline2 (reused stages)
	pipeline2.AddStage(multiplyBy2)
	pipeline2.AddStage(fgl_stage.NewStage[float64](total))

	// Log starting the pipeline
	time1 := time.Now().UTC()
	fmt.Printf("Starting pipelines at %s\n", time1)

	// Kick off the pipelines in parallel
	job := fgl_pipeline.PipelineJob[float64]{}

	job.AddPipeline(pipeline)
	job.AddPipeline(pipeline2)

	// Run all pipelines in parallel
	job.RunPipelinesInParallel(context.Background())

	// // Start pipeline in a goroutine
	// go func() {
	// 	err := pipeline.RunPipeline(context.Background())
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// }()

	// // Log starting the pipeline2
	// time2 := time.Now().UTC()
	// fmt.Printf("Starting pipeline2 at %s\n", time2.String())

	// // Start pipeline2 in a goroutine
	// go func() {
	// 	err := pipeline2.RunPipeline(context.Background())
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// }()

	// Send input values in a goroutine
	go func() {
		intialValue := 10.0

		// Send input values to pipeline1
		for i := 0; i < 10000000; i++ {
			in <- intialValue * 10.0
			intialValue += 1.0
		}

		close(in)
	}()

	go func() {
		intialValue := 10.0

		// Send input values to pipeline2
		for i := 0; i < 1000000; i++ {
			in2 <- intialValue / 10.0
			intialValue -= 10.0
		}

		close(in2)
	}()

	go func() {
		for result := range out {
			fmt.Printf("Final result: %v\n", result)
		}

		wg.Done()
	}()

	wg.Wait()

	go func() {
		for result := range out2 {
			fmt.Printf("Final result2: %v\n", result)
		}

		wg2.Done()
	}()

	wg2.Wait()
}
