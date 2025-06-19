package main

import (
	"context"
	"fastgoline/models"
	"fmt"
	"math"
	"sync"
)

func main() {
	// Create a new pipeline
	pipeline := models.NewPipeline[float64]()

	// Create input and output channels
	in := make(chan float64)
	out := make(chan float64, 3)

	wg := sync.WaitGroup{}
	wg.Add(1)

	// Define stage functions
	multiplyBy2 := models.StageTransformFunction[float64](func(ctx context.Context, in <-chan float64, out chan<- float64) error {
		for value := range in {
			fmt.Printf("Stage 2 output: %v\n", value*2)
			out <- value * 2
		}

		return nil
	})

	percentage := models.StageTransformFunction[float64](func(ctx context.Context, in <-chan float64, out chan<- float64) error {
		for value := range in {
			fmt.Printf("Stage 2 output: %v\n", value/100)
			out <- value / 100
		}

		return nil
	})

	total := models.StageTransformFunction[float64](func(ctx context.Context, in <-chan float64, out chan<- float64) error {
		var total float64
		for value := range in {
			total += value
		}

		fmt.Printf("Stage 3 output: %v\n", math.Floor(total*100)/100)

		out <- math.Floor(total*100) / 100
		return nil
	})

	// Add stages to the pipeline
	pipeline.AddStage(models.NewStage[float64](multiplyBy2))
	pipeline.AddStage(models.NewStage[float64](percentage))
	pipeline.AddStage(models.NewStage[float64](total))

	// Start pipeline in a goroutine
	go func() {
		err := pipeline.RunPipeline(context.Background(), in, out)
		if err != nil {
			panic(err)
		}
	}()

	// Send input values in a goroutine
	go func() {
		in <- 10.0
		in <- 20.0
		in <- 30.0

		close(in)
	}()

	go func() {
		for result := range out {
			fmt.Printf("Final result: %v\n", result)
		}
		wg.Done()
	}()

	wg.Wait()
}
