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

```go
package main

import (
	"context"
	fglpipeline "fastgoline/fgl/pipeline"
	fglstage "fastgoline/fgl/stage"
	"fmt"
	"math"
	"sync"
	"time"
)

func main() {
	// Channels for pipeline 1
	in := make(chan float64)
	out := make(chan float64, 3)

	pipeline1 := fglpipeline.NewPipeline[float64](in, out)

	// Channels for pipeline 2
	in2 := make(chan float64)
	out2 := make(chan float64, 3)

	pipeline2 := fglpipeline.NewPipeline[float64](in2, out2)

	// Define reusable stage using StageTransformFunction
	// This is used when you want full control over input and output channels,
	// such as for aggregations (sum, average, filtering, etc.)
	total := fglstage.StageTransformFunction[float64](func(ctx context.Context, in <-chan float64, out chan<- float64) error {
		var total float64
		for value := range in {
			total += value
		}
		out <- math.Floor(total*100) / 100
		return nil
	})

	// For simpler transformations (like activation functions or single-value transforms),
	// you can use NewStageFunction which abstracts channel handling.
	// These are suitable for stateless functions like scaling, normalization, etc.
	multiplyBy2 := fglstage.NewStageFunction[float64](func(value float64) float64 {
		return value * 2
	})

	percentage := fglstage.NewStageFunction[float64](func(value float64) float64 {
		return value / 100
	})

	// Add stages to each pipeline
	pipeline1.AddStage(multiplyBy2)
	pipeline1.AddStage(fglstage.NewStage[float64](total))
	pipeline1.AddStage(percentage)

	pipeline2.AddStage(multiplyBy2)
	pipeline2.AddStage(fglstage.NewStage[float64](total))

	// Run pipelines concurrently
	job := fglpipeline.PipelineJob[float64]{}

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
			fmt.Printf("Final result (pipeline1): %v\n", result)
		}

		wg.Done()
	}()

	go func() {
		for result := range out2 {
			fmt.Printf("Final result (pipeline2): %v\n", result)
		}

		wg.Done()
	}()

	wg.Wait()
}
```

## 🧪 Example Output

```txt
Starting pipelines at 2025-06-22 17:23:14.494245 +0000 UTC
Running 2 pipelines in parallel...
Starting pipeline 0000XSNJG0MQJHBF4QX1EFD6Y3 at 2025-06-22 17:23:14.494407 +0000 UTC
Starting pipeline 0000XSNJG0MQJHBF4QX1EFD6Y3 at 2025-06-22 17:23:14.494405 +0000 UTC
Final result (pipeline2): -9.99997e+11
Finished Pipeline 0000XSNJG0MQJHBF4QX1EFD6Y3 in 243 ms
Finished Pipeline 0000XSNJG0MQJHBF4QX1EFD6Y3 in 2812 ms
Final result (pipeline1): 1.0000019e+13
```

This demonstrates two pipelines running concurrently, with different data and total execution times.

## 📄 License

This project is licensed under the [MIT License](./LICENSE).
