# fastgoline

**FastGoLine** is a high-performance data pipeline processing library built in Go.

---

## ✨ Features

* ⚙️ **Generic Type Support** – Works with any data type using Go generics.

* 🔄 **Concurrent Stage Execution** – Each pipeline stage runs in its own goroutine.

* 🔁 **Stage Reusability** – Reuse transformation logic across pipelines.

* 🔗 **Pipeline Composition** – Chain multiple stages to transform input to output.

* 🧠 **Deadlock-Safe** – Designed to prevent blocking and goroutine leaks.

* 🧪 **Multi-Pipeline Execution** – Run multiple pipelines in parallel with shared or unique stages.

* 🪪 **UUID-based Identification** – Every pipeline and stage have their unique ID for tracing.


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
	"fastgoline/fgl"
	"fmt"
	"math"
	"sync"
	"time"
)

func main() {
	// Channels for pipeline 1
	in := make(chan float64)
	out := make(chan float64, 3)

	pipeline1 := fgl.NewPipeline[float64](in, out)

	// Channels for pipeline 2
	in2 := make(chan float64)
	out2 := make(chan float64, 3)

	pipeline2 := fgl.NewPipeline[float64](in2, out2)

	// Shared stage functions
	multiplyBy2 := fgl.StageTransformFunction[float64](func(ctx context.Context, in <-chan float64, out chan<- float64) error {
		for value := range in {
			out <- value * 2
		}

		return nil
	})

	percentage := fgl.StageTransformFunction[float64](func(ctx context.Context, in <-chan float64, out chan<- float64) error {
		for value := range in {
			out <- value / 100
		}

		return nil
	})

	total := fgl.StageTransformFunction[float64](func(ctx context.Context, in <-chan float64, out chan<- float64) error {
		var total float64
		for value := range in {
			total += value
		}
		out <- math.Floor(total*100) / 100
		return nil
	})

	// Add stages to each pipeline
	pipeline1.AddStage(fgl.NewStage[float64](multiplyBy2))
	pipeline1.AddStage(fgl.NewStage[float64](percentage))
	pipeline1.AddStage(fgl.NewStage[float64](total))

	pipeline2.AddStage(fgl.NewStage[float64](multiplyBy2))
	pipeline2.AddStage(fgl.NewStage[float64](total))

	// Run pipelines concurrently
	job := fgl.PipelineJob[float64]{}
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
		// Send input values to pipeline1
		for i := 0; i < 10000000; i++ {
			in <- intialValue * 10.0
			intialValue += 1.0
		}

		close(in)
	}()

	go func() {
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
Starting pipelines at 2025-06-21 21:58:49.224793 +0000 UTC
Running 2 pipelines in parallel...
Starting pipeline 0000XSNJG0MQJHBF4QX1EFD6Y3 at 2025-06-21 21:58:49.224882 +0000 UTC
Starting pipeline 0000XSNJG0MQJHBF4QX1EFD6Y3 at 2025-06-21 21:58:49.224883 +0000 UTC
Finished Pipeline 0000XSNJG0MQJHBF4QX1EFD6Y3 in 243 ms
Finished Pipeline 0000XSNJG0MQJHBF4QX1EFD6Y3 in 4626 ms
Final result: 1.0000019e+13
Final result2: -9.99997e+11
```

This demonstrates two pipelines running concurrently, with different data and total execution times.

## 📄 License

This project is licensed under the [MIT License](./LICENSE).
