package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	fgl_config "github.com/hkashwinkashyap/fastgoline/fgl/config"
	fgl_metadata "github.com/hkashwinkashyap/fastgoline/fgl/metadata"
	fgl_pipeline "github.com/hkashwinkashyap/fastgoline/fgl/pipeline"
	fgl_stage "github.com/hkashwinkashyap/fastgoline/fgl/stage"
	fgl_util "github.com/hkashwinkashyap/fastgoline/util"
)

// logMemStats logs memory stats to the log file
func logMemStats(f *os.File) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	stats := fmt.Sprintf(
		"Memory Stats: Alloc = %v MiB, TotalAlloc = %v MiB, Sys = %v MiB, NumGC = %v\n",
		fgl_util.BytesToMB(m.Alloc), fgl_util.BytesToMB(m.TotalAlloc), fgl_util.BytesToMB(m.Sys), m.NumGC,
	)
	_, err := f.WriteString(stats)
	if err != nil {
		fmt.Println("Error writing memory stats:", err)
	}
}

func main() {
	// Create log file
	f, err := os.Create(fmt.Sprintf("fastgoline_test_%s.log", time.Now().Format("2006-01-02 15:04:05")))
	if err != nil {
		fmt.Println("Error creating log file:", err)
		return
	}
	defer func() {
		err := f.Close()
		if err != nil {
			fmt.Println("Error closing log file:", err)
		}
	}()

	// Start periodic memory stats logging every 100 milliseconds in background
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for range ticker.C {
			logMemStats(f)
		}
	}()

	// Channels for pipeline 1
	in := make(chan fgl_metadata.InputMetadata[float64])
	out := make(chan fgl_metadata.OutputMetadata[float64])

	// Initialise config
	config := fgl_config.InitialiseConfig()
	// Change log level as required
	config.LogLevel = fgl_config.LogLevelError
	// config.MaxWorkers = 64

	// Write config to log
	line1 := fmt.Sprintf("FastGoLine Test Log with `max_workers` set to %d with CPUs: %d\n", config.MaxWorkers, runtime.NumCPU())
	_, err = f.WriteString(line1)
	if err != nil {
		fmt.Println("Error writing to log file:", err)
		return
	}

	pipeline1 := fgl_pipeline.NewPipeline[float64](in, out, config)

	// Channels for pipeline 2
	in2 := make(chan fgl_metadata.InputMetadata[float64])
	out2 := make(chan fgl_metadata.OutputMetadata[float64])

	pipeline2 := fgl_pipeline.NewPipeline[float64](in2, out2, config)

	// Define reusable stage using StageTransformFunction
	// This is used when you want full control over input and output,
	// such as for aggregations (sum, average, filtering, etc.)
	total := fgl_stage.StageTransformFunction[float64](func(ctx context.Context, in fgl_metadata.InputMetadata[float64]) fgl_metadata.OutputMetadata[float64] {
		var total float64

		total += in.Value

		return fgl_metadata.OutputMetadata[float64]{
			InitialInput: fgl_metadata.InitialInput[float64]{
				Id:        in.InitialInput.Id,
				Value:     in.InitialInput.Value,
				InputTime: in.InitialInput.InputTime,
			},
			OutputValue: total,
			OutputTime:  time.Now().UTC(),
			Duration:    time.Since(in.InputTime).Abs().Nanoseconds(),
			Err:         nil,
		}
	})

	// For simpler transformations (like activation functions or single-value transforms),
	// you can use NewStageFunction which abstracts channel handling.
	// These are suitable for stateless functions like scaling, normalization, etc.
	multiplyBy2 := fgl_stage.NewStageFunction[float64](func(value float64) (float64, error) {
		return value * 2, nil
	})

	percentage := fgl_stage.NewStageFunction[float64](func(value float64) (float64, error) {
		return value / 100, nil
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
	line2 := fmt.Sprintf("Starting pipelines at %s\n", startTime)
	_, err = f.WriteString(line2)
	if err != nil {
		fmt.Println("Error writing to log file:", err)
		return
	}

	// Launch all pipelines
	job.RunPipelinesInParallel(context.Background())

	// Pass in the input to the channels
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		intialValue := 0.0
		defer close(in)

		// Send input values to pipeline1
		for i := 0; i < 1000; i++ {
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
		for i := 0; i < 1000; i++ {
			in2 <- fgl_metadata.NewInputMetadata(intialValue / 10.0)
			// Add delay if required
			// time.Sleep(time.Second * 1)
			intialValue -= 10.0
		}
	}()

	go func() {
		counter := 0
		for range out {
			counter++

			if counter == 1000 {
				break
			}
		}

		line1 := fmt.Sprintf("Pipeline1 completed: %d items\n", counter)
		line2 := fmt.Sprintf("Pipeline1 completed in %+vms\n", time.Since(startTime).Milliseconds())

		_, err = f.WriteString(line1 + line2)
		if err != nil {
			fmt.Println("Error writing to log file:", err)
			return
		}
		wg.Done()
	}()

	go func() {
		counter := 0
		for range out2 {
			counter++

			if counter == 1000 {
				break
			}
		}

		line1 := fmt.Sprintf("Pipeline2 completed: %d items\n", counter)
		line2 := fmt.Sprintf("Pipeline2 completed in %+vms\n", time.Since(startTime).Milliseconds())

		_, err = f.WriteString(line1 + line2)
		if err != nil {
			fmt.Println("Error writing to log file:", err)
			return
		}
		wg.Done()
	}()

	wg.Wait()
	endTime := time.Now().UTC()
	line3 := fmt.Sprintf("All pipelines completed at %s\n", endTime)
	line4 := fmt.Sprintf("Total time: %+v ms\n", endTime.Sub(startTime).Milliseconds())

	_, err = f.WriteString(line3 + line4)
	if err != nil {
		fmt.Println("Error writing to log file:", err)
		return
	}
}
