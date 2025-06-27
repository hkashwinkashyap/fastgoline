# fastgoline

**FastGoLine** is a high-performance data pipeline processing library built in Go.

---

## ✨ Features

* **Generic Type Support** – Works with any data type using Go generics.

* **Concurrent Stage Execution** – Each pipeline stage runs in its own goroutine.

* **Stage Reusability** – Reuse transformation logic across pipelines.

* **Pipeline Composition** – Chain multiple stages to transform input to output.

* **Deadlock-Safe** – Designed to prevent blocking and goroutine leaks.

* **Multi-Pipeline Execution** – Run multiple pipelines in parallel with shared or unique stages.

* **UUID-based Identification** – Every pipeline and stage have their unique ID for tracing.


## 📦 Installation

```bash
go get github.com/hkashwinkashyap/fastgoline
```

## 🚀 Usage

Here’s a complete example of how to define and run multiple pipelines concurrently:

NOTE: The environment variables can be set on the machine and when `fgl_config.InitialiseConfig()` is called, the values will be pulled in.

```go
package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	fgl_config "github.com/hkashwinkashyap/fastgoline/fgl/config"
	fgl_metadata "github.com/hkashwinkashyap/fastgoline/fgl/metadata"
	fgl_pipeline "github.com/hkashwinkashyap/fastgoline/fgl/pipeline"
	fgl_stage "github.com/hkashwinkashyap/fastgoline/fgl/stage"
)

func main() {
	// Channels for pipeline 1
	in := make(chan fgl_metadata.InputMetadata[float64])
	out := make(chan fgl_metadata.OutputMetadata[float64])

	// Initialise config
	config := fgl_config.InitialiseConfig()
	// Change log level as required
	// config.LogLevel = fgl_config.LogLevelError

	pipeline1 := fgl_pipeline.NewPipeline[float64](in, out, config)

	// Channels for pipeline 2
	in2 := make(chan fgl_metadata.InputMetadata[float64])
	out2 := make(chan fgl_metadata.OutputMetadata[float64])

	pipeline2 := fgl_pipeline.NewPipeline[float64](in2, out2, config)

	// Define reusable stage using StageTransformFunction
	// This is used when you want full control over input and output channels,
	// such as for aggregations (sum, average, filtering, etc.)
	total := fgl_stage.StageTransformFunction[float64](func(ctx context.Context, in chan fgl_metadata.InputMetadata[float64], out chan fgl_metadata.OutputMetadata[float64]) error {
		var total float64

		for value := range in {
			total += value.Value

			out <- fgl_metadata.OutputMetadata[float64]{
				InitialInput: fgl_metadata.InitialInput[float64]{
					Id:        value.InitialInput.Id,
					Value:     value.InitialInput.Value,
					InputTime: value.InitialInput.InputTime,
				},
				OutputValue: total,
				OutputTime:  time.Now().UTC(),
				Duration:    time.Since(value.InputTime).Abs().Nanoseconds(),
				Err:         nil,
			}
		}

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
	pipeline1.AddStage(percentage)
	pipeline1.AddStage(multiplyBy2)
	pipeline1.AddStage(multiplyBy2)
	pipeline1.AddStage(multiplyBy2)
	pipeline1.AddStage(percentage)

	pipeline2.AddStage(multiplyBy2)
	pipeline2.AddStage(fgl_stage.NewStage[float64](total))

	// Run pipelines concurrently
	job := fgl_pipeline.NewPipelineJob[float64](config)

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
		intialValue := 0.0
		defer close(in)

		// Send input values to pipeline1
		for i := 0; i < 100000; i++ {
			in <- fgl_metadata.NewInputMetadata(intialValue * 10.0)
			// Add delay if required
			// time.Sleep(time.Second * 1)
			intialValue += 10.0
		}
	}()

	go func() {
		intialValue := 10.0
		defer close(in2)

		// Send input values to pipeline2
		for i := 0; i < 100; i++ {
			in2 <- fgl_metadata.NewInputMetadata(intialValue / 10.0)
			// Add delay if required
			// time.Sleep(time.Second * 1)
			intialValue -= 10.0
		}
	}()

	go func() {
		counter := 0
		for result := range out {
			fmt.Printf("Final result (pipeline1): %+v\n", result)
			counter++

			if counter == 100000 {
				break
			}
		}

		wg.Done()
	}()

	go func() {
		counter := 0
		for result := range out2 {
			fmt.Printf("Final result (pipeline2): %+v\n", result)
			counter++

			if counter == 100 {
				break
			}
		}

		wg.Done()
	}()

	wg.Wait()
}
```

## 📄 License

This project is licensed under the [MIT License](./LICENSE).
