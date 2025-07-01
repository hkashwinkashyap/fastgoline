package fgl_pipeline

import (
	"context"
	"fmt"
	"time"

	fgl_config "github.com/hkashwinkashyap/fastgoline/fgl/config"
	fgl_metadata "github.com/hkashwinkashyap/fastgoline/fgl/metadata"
	fgl_stage "github.com/hkashwinkashyap/fastgoline/fgl/stage"

	fgl_util "github.com/hkashwinkashyap/fastgoline/util"
)

// Pipeline holds a sequence of stages to process data.
// Each stage runs concurrently, passing data along the pipeline.
type Pipeline[T any] struct {
	id            string
	Stages        []fgl_stage.Stage[T]
	InputChannel  chan fgl_metadata.InputMetadata[T]
	OutputChannel chan fgl_metadata.OutputMetadata[T]
	Done          chan struct{}
	StartedAt     time.Time
	CompletedAt   time.Time
	Config        fgl_config.Config
}

// GetID returns the ID of the pipeline.
func (pipeline *Pipeline[T]) GetID() string {
	return pipeline.id
}

// NewPipeline creates an empty pipeline instance.
func NewPipeline[T any](in chan fgl_metadata.InputMetadata[T], out chan fgl_metadata.OutputMetadata[T], config *fgl_config.Config) *Pipeline[T] {
	// Generate a unique id
	id := fgl_util.GenerateUUID()

	if config == nil {
		config = fgl_config.InitialiseConfig()
	}

	// Return a new pipeline instance
	return &Pipeline[T]{id: id,
		Stages:        nil,
		InputChannel:  in,
		OutputChannel: out,
		Done:          make(chan struct{}),
		StartedAt:     time.Now().UTC(),
		Config:        *config,
	}
}

// worker goroutine
// It processes the pipeline stages concurrently
func (pipeline *Pipeline[T]) worker(ctx context.Context, inputQueue chan fgl_metadata.InputMetadata[T], workerMetadata *fgl_metadata.WorkerMetadata[T]) {
	semaphore := make(chan struct{}, pipeline.Config.MaxWorkers)

	// Loop through each input coming from the input and process it through the pipeline
	for inputValue := range inputQueue {
		// Wait until a worker slot is available
		semaphore <- struct{}{}

		workerMetadata.IncrementActiveWorkers()

		// if workerMetadata.GetActiveWorkers() >= pipeline.Config.MaxWorkers {
		// 	fmt.Printf("WARN: Pipeline %s is at max workers. Waiting for a worker to be freed up...\n", pipeline.id)
		// }

		go func(fgl_metadata.InputMetadata[T]) {
			now := time.Now().UTC()
			inputValue.InitialInput = &fgl_metadata.InitialInput[T]{
				Id:        fgl_util.GenerateUUID(),
				Value:     inputValue.Value,
				InputTime: now,
			}

			if pipeline.Config.LogLevel == fgl_config.LogLevelDebug {
				fmt.Printf("TRACE: Starting pipeline %s with input - {Id: %s, Value: %+v, InputTime: %s}\n", pipeline.id, inputValue.InitialInput.Id, inputValue.InitialInput.Value, inputValue.InitialInput.InputTime)
			}

			// Current input
			currentIn := inputValue

			// Current output is the intermediate output used to chain the stages thorughout the pipeline
			currentOut := fgl_metadata.OutputMetadata[T]{}

			// Pipe in the inputValue to the first stage
			for index, stage := range pipeline.Stages {
				// Make a note of timestamp when the stage is started to calculate the duration or time taken to process that stage
				startedStageAt := time.Now().UTC()

				currentOut = stage.TransformFunction(ctx, currentIn)
				if currentOut.Err != nil {
					ctx.Done()
					fmt.Printf("ERR: Failed to process stage number %d {%+v} of pipeline %s: %s\n", index, stage, pipeline.id, currentOut.Err.Error())
					panic(currentOut.Err)
				}

				if pipeline.Config.LogLevel == fgl_config.LogLevelDebug {
					fmt.Printf("TRACE: Processed stage %s in %v nanoseconds\n", stage.GetID(), time.Since(startedStageAt).Abs().Nanoseconds())
				}

				// Current output becomes the next input
				currentIn = convertCurrentOutToNextIn(currentOut)
			}

			// Mark the pipeline as completed
			pipeline.CompletedAt = time.Now().UTC()
			pipeline.OutputChannel <- currentOut

			// Release the worker
			<-semaphore
			workerMetadata.DecrementActiveWorkers()

			if pipeline.Config.LogLevel == fgl_config.LogLevelInfo {
				fmt.Printf("TRACE: Finished Pipeline %s in %+v ms with output value %+v\n", pipeline.GetID(), time.Now().UTC().Sub(now).Milliseconds(), currentOut)
			}

			if ctx.Err() != nil {
				fmt.Printf("ERROR: Pipeline %s failed with error %v\n", pipeline.GetID(), ctx.Err())
				panic(ctx.Err())
			}

			pipeline.Done <- struct{}{}
		}(inputValue)
	}
}

// RunPipeline runs the pipeline to process data.
// It returns an error if any stage fails.
func (pipeline *Pipeline[T]) RunPipeline(ctx context.Context) {
	workerMetadata := fgl_metadata.NewWorkerMetadata[T]()
	// maxWorkers := pipeline.Config.MaxWorkers
	// maxMemoryMB := pipeline.Config.MaxMemoryMB

	inputQueue := make(chan fgl_metadata.InputMetadata[T])

	// Spawn fixed number of goroutines
	go pipeline.worker(ctx, inputQueue, workerMetadata)

	// Send input values to the input
	go func() {
		for input := range pipeline.InputChannel {
			select {
			case inputQueue <- input:
			case <-ctx.Done():
				return
			}
		}

		close(inputQueue)
	}()
}

// convertCurrentOutToNextIn converts output metadata to input metadata
// This is used to pipe the output of a stage to the input of the next intermediate stage
func convertCurrentOutToNextIn[T any](currentOut fgl_metadata.OutputMetadata[T]) fgl_metadata.InputMetadata[T] {
	nextIn := fgl_metadata.InputMetadata[T]{
		InitialInput: &currentOut.InitialInput,
		Value:        currentOut.OutputValue,
		InputTime:    currentOut.InitialInput.InputTime,
	}

	// Return the next input
	return nextIn
}

// AddStage appends a processing stage to the pipeline.
func (pipeline *Pipeline[T]) AddStage(stage fgl_stage.Stage[T]) {
	pipeline.Stages = append(pipeline.Stages, stage)
}

// AddStages appends multiple processing stages to the pipeline.
func (pipeline *Pipeline[T]) AddStages(stages []fgl_stage.Stage[T]) {
	pipeline.Stages = append(pipeline.Stages, stages...)
}

// RemoveStage removes a processing stage from the pipeline.
func (pipeline *Pipeline[T]) RemoveStage(stage fgl_stage.Stage[T]) {
	for index, stageItem := range pipeline.Stages {
		if stageItem.GetID() == stage.GetID() {
			pipeline.Stages = append(pipeline.Stages[:index], pipeline.Stages[index+1:]...)
			break
		}
	}
}
