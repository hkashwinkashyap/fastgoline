package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"

	_ "net/http/pprof"

	fgl_config "github.com/hkashwinkashyap/fastgoline/fgl/config"
	fgl_metadata "github.com/hkashwinkashyap/fastgoline/fgl/metadata"
	fgl_pipeline "github.com/hkashwinkashyap/fastgoline/fgl/pipeline"
	fgl_stage "github.com/hkashwinkashyap/fastgoline/fgl/stage"
	fgl_util "github.com/hkashwinkashyap/fastgoline/util"
)

const (
	numPipelines      = 10
	inputsPerPipeline = 2000000
	logInterval       = 100000
)

// logMemStats logs memory stats to the log file
func logMemStats(f *os.File, counter int, pipelineNumber int, duration int64) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	stats := fmt.Sprintf(
		"Memory Stats: Alloc = %v MiB, TotalAlloc = %v MiB, Sys = %v MiB, NumGC = %v, Goroutines = %d, Counter = %v, Pipeline = %v took %vms\n",
		fgl_util.BytesToMB(m.Alloc), fgl_util.BytesToMB(m.TotalAlloc), fgl_util.BytesToMB(m.Sys),
		m.NumGC, runtime.NumGoroutine(), counter, pipelineNumber, duration,
	)
	_, err := f.WriteString(stats)
	if err != nil {
		fmt.Println("Error writing memory stats:", err)
	}
	fmt.Println(stats)
}

func main() {
	// Start pprof
	go func() {
		fmt.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	runtime.GOMAXPROCS(runtime.NumCPU())

	// Create log file
	f, err := os.Create(fmt.Sprintf("fastgoline_test_%s.log", time.Now().Format("2006-01-02 15:04:05")))
	if err != nil {
		panic(err)
	}

	defer func() {
		err = f.Close()
		if err != nil {
			panic(err)
		}
	}()

	// Initialise config
	config := fgl_config.InitialiseConfig()
	config.LogLevel = fgl_config.LogLevelError

	logLine := fmt.Sprintf("FastGoLine Test Log: max_workers=%d, CPUs=%d, pipelines=%d\n", config.MaxWorkers, runtime.NumCPU(), numPipelines)
	_, err = f.WriteString(logLine)
	if err != nil {
		panic(err)
	}

	// Define stages
	multiplyBy2 := fgl_stage.NewStageFunction[float64](func(value float64) (float64, error) {
		return value * 2, nil
	})
	percentage := fgl_stage.NewStageFunction[float64](func(value float64) (float64, error) {
		return value / 100, nil
	})

	// Pipeline inputs and outputs
	inChans := make([]chan fgl_metadata.InputMetadata[float64], numPipelines)
	outChans := make([]chan fgl_metadata.OutputMetadata[float64], numPipelines)

	job := fgl_pipeline.NewPipelineJob[float64](config)

	// Setup pipelines
	for i := 0; i < numPipelines; i++ {
		inChans[i] = make(chan fgl_metadata.InputMetadata[float64], 1000)
		outChans[i] = make(chan fgl_metadata.OutputMetadata[float64], 1000)

		p := fgl_pipeline.NewPipeline[float64](inChans[i], outChans[i], config)
		p.AddStage(multiplyBy2)
		p.AddStage(percentage)
		p.AddStage(multiplyBy2)
		p.AddStage(multiplyBy2)
		p.AddStage(multiplyBy2)
		p.AddStage(percentage)

		job.AddPipeline(p)
	}

	startTime := time.Now().UTC()
	startingLog := fmt.Sprintf("Starting stress test at %s\n", startTime.Format(time.RFC3339))
	_, err = f.WriteString(startingLog)
	if err != nil {
		panic(err)
	}

	fmt.Println(startingLog)

	// Start pipeline processing
	job.RunPipelinesInParallel(context.Background())

	var wg sync.WaitGroup

	// Input goroutines
	for i := 0; i < numPipelines; i++ {
		wg.Add(1)
		go func(i int) {
			defer close(inChans[i]) // Close input after sending all items
			value := float64(i * 100)
			for j := 0; j < inputsPerPipeline; j++ {
				inChans[i] <- fgl_metadata.NewInputMetadata(value)
				value += 1.1
			}
			wg.Done()
		}(i)
	}

	// Wait for all inputs to be sent before handling output close
	go func() {
		wg.Wait()
		// Optional: add small delay to let processing complete
		time.Sleep(2 * time.Second)
		for _, out := range outChans {
			close(out)
		}
	}()

	// Output collectors
	var outputWg sync.WaitGroup
	for i := 0; i < numPipelines; i++ {
		outputWg.Add(1)
		go func(i int) {
			defer outputWg.Done()
			counter := 0
			var lastOutput fgl_metadata.OutputMetadata[float64]
			for output := range outChans[i] {
				counter++
				lastOutput = output
				if counter%logInterval == 0 {
					logMemStats(f, counter, i+1, time.Since(startTime).Milliseconds())
				}
			}
			completedLog := fmt.Sprintf("Pipeline %d completed: %d items in %d ms. Last Output: %+v\n",
				i+1, counter, time.Since(startTime).Milliseconds(), lastOutput)
			_, err = f.WriteString(completedLog)
			if err != nil {
				panic(err)
			}
		}(i)
	}

	outputWg.Wait()
	endTime := time.Now().UTC()
	allPipelinesCompleted := fmt.Sprintf("All pipelines completed at %s\n", endTime.Format(time.RFC3339))
	_, err = f.WriteString(allPipelinesCompleted)
	if err != nil {
		panic(err)
	}

	totalElapsedTime := fmt.Sprintf("Total elapsed time: %d ms\n", endTime.Sub(startTime).Milliseconds())
	_, err = f.WriteString(totalElapsedTime)
	if err != nil {
		panic(err)
	}
}
