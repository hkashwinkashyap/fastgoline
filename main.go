package main

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/hkashwinkashyap/fastgoline/fgl/metadata"
	fgl_pipeline "github.com/hkashwinkashyap/fastgoline/fgl/pipeline"
	fgl_stage "github.com/hkashwinkashyap/fastgoline/fgl/stage"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(2)

	// Pipeline 1 setup
	in1 := make(chan metadata.InputMetadata[float64])
	out1 := make(chan metadata.OutputMetadata[float64], 100)
	pipeline1 := fgl_pipeline.NewPipeline[float64](in1, out1, false)

	// Pipeline 2 setup
	in2 := make(chan metadata.InputMetadata[float64])
	out2 := make(chan metadata.OutputMetadata[float64], 100)
	pipeline2 := fgl_pipeline.NewPipeline[float64](in2, out2, false)

	// Define simple processing stages
	multiplyBy2 := fgl_stage.NewStageFunction[float64](func(v float64) float64 {
		return v * 2
	})

	percentage := fgl_stage.NewStageFunction[float64](func(v float64) float64 {
		return v / 100
	})

	// Define a custom total stage (accumulate all values)
	total := fgl_stage.NewStage[float64](func(ctx context.Context, in <-chan metadata.InputMetadata[float64], out chan<- metadata.OutputMetadata[float64]) error {
		var sum float64
		var id string
		var start time.Time

		for v := range in {
			if id == "" {
				id = v.Id
				start = v.InputTime
			}
			sum += v.Value
		}

		out <- metadata.OutputMetadata[float64]{
			InputID:    id,
			Value:      math.Floor(sum*100) / 100,
			InputTime:  start,
			OutputTime: time.Now(),
			Duration:   time.Since(start).Milliseconds(),
		}
		return nil
	})

	// Add stages to pipeline1 and pipeline2
	pipeline1.AddStage(multiplyBy2)
	pipeline1.AddStage(percentage)
	pipeline1.AddStage(total)

	pipeline2.AddStage(percentage)
	pipeline2.AddStage(multiplyBy2)
	pipeline2.AddStage(total)

	// Run pipeline1
	go func() {
		if err := pipeline1.RunPipeline(context.Background()); err != nil {
			panic(err)
		}
		wg.Done()
	}()

	// Run pipeline2
	go func() {
		if err := pipeline2.RunPipeline(context.Background()); err != nil {
			panic(err)
		}
		wg.Done()
	}()

	// Send input data to pipeline1
	go func() {
		for i := 1.0; i <= 10000; i++ {
			in1 <- metadata.NewInputMetadata(i)
		}
		close(in1)
	}()

	// Send input data to pipeline2
	go func() {
		for i := 10000.0; i >= 1; i-- {
			in2 <- metadata.NewInputMetadata(i)
		}
		close(in2)
	}()

	// Read output from pipeline1
	go func() {
		for result := range out1 {
			fmt.Printf("Pipeline1 Output: %+v\n", result)
		}
	}()

	// Read output from pipeline2
	go func() {
		for result := range out2 {
			fmt.Printf("Pipeline2 Output: %+v\n", result)
		}
	}()

	wg.Wait()
}
