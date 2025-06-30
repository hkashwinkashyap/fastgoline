package fgl_pipeline

import (
	"context"
	"fmt"
	"sync"
	"time"

	fgl_config "github.com/hkashwinkashyap/fastgoline/fgl/config"
	fgl_metadata "github.com/hkashwinkashyap/fastgoline/fgl/metadata"
	fgl_stage "github.com/hkashwinkashyap/fastgoline/fgl/stage"

	fgl_util "github.com/hkashwinkashyap/fastgoline/util"
)

// Pipeline holds a sequence of stages to process data.
// Each stage runs concurrently, passing data along channels.
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
		Stages: nil, InputChannel: in,
		OutputChannel: out,
		Done:          make(chan struct{}),
		StartedAt:     time.Now().UTC(),
		Config:        *config,
	}
}

// worker goroutine
// It processes the pipeline stages concurrently
func (pipeline *Pipeline[T]) worker(ctx context.Context, inputQueue chan fgl_metadata.InputMetadata[T]) {
	// Loop through each input coming from the input channel and process it through the pipeline
	for inputValue := range inputQueue {
		go func(inputValue fgl_metadata.InputMetadata[T]) {
			now := time.Now().UTC()
			inputValue.InitialInput = &fgl_metadata.InitialInput[T]{
				Id:        fgl_util.GenerateUUID(),
				Value:     inputValue.Value,
				InputTime: now,
			}

			if pipeline.Config.LogLevel == "DEBUG" {
				fmt.Printf("TRACE: Starting pipeline %s with input - {Id: %s, Value: %+v, InputTime: %s}\n", pipeline.id, inputValue.InitialInput.Id, inputValue.InitialInput.Value, inputValue.InitialInput.InputTime)
			}

			var err error
			var wg sync.WaitGroup

			// Set the number of goroutines to the number of stages
			wg.Add(len(pipeline.Stages))

			// Pipe in the inputValue to the first stage
			currentIn := make(chan fgl_metadata.InputMetadata[T], 1)
			currentIn <- inputValue

			for index, stage := range pipeline.Stages {
				// Current output channel is the intermediate output channel used to chain the stages thorughout the pipeline
				currentOut := make(chan fgl_metadata.OutputMetadata[T], 1)

				isFinal := index == len(pipeline.Stages)-1

				if isFinal {
					// Final stage
					// Pipe the final output to the provided output channel of the pipeline
					err = processStageTransform(ctx, stage, currentIn, pipeline.OutputChannel, &wg)
					if err != nil {
						ctx.Done()
						fmt.Printf("ERR: Failed to process final stage of pipeline %s: %s\n", pipeline.id, err.Error())
						panic(err)
					}
				} else {
					// Make a note of timestamp when the stage is started to calculate the duration or time taken to process that stage
					startedStageAt := time.Now().UTC()

					// Intermediate stage
					// Pipe the current output to the next input channel
					err = processStageTransform(ctx, stage, currentIn, currentOut, &wg)
					if err != nil {
						ctx.Done()
						fmt.Printf("ERR: Failed to process intermediate stage of pipeline %s: %s\n", pipeline.id, err.Error())
						panic(err)
					}

					if pipeline.Config.LogLevel == "DEBUG" {
						fmt.Printf("TRACE: Processed stage %s in %v nanoseconds\n", stage.GetID(), time.Since(startedStageAt).Abs().Nanoseconds())
					}

					// Current output channel becomes the next input channel
					currentIn = convertCurrentOutToNextIn(currentOut)
				}
			}

			// Wait for all stages to finish
			wg.Wait()

			// Mark the pipeline as completed
			pipeline.CompletedAt = time.Now().UTC()
			pipeline.Done <- struct{}{}

			if pipeline.Config.LogLevel == "DEBUG" {
				fmt.Printf("TRACE: Finished Pipeline %s in %v ms with input value %+v\n", pipeline.GetID(), time.Now().UTC().Sub(now).Milliseconds(), inputValue)
			}

			if ctx.Err() != nil {
				fmt.Printf("ERROR: Pipeline %s failed with error %v\n", pipeline.GetID(), ctx.Err())
				panic(ctx.Err())
			}
		}(inputValue)
	}
}

// RunPipeline runs the pipeline to process data.
// It returns an error if any stage fails.
func (pipeline *Pipeline[T]) RunPipeline(ctx context.Context) {
	maxWorkers := pipeline.Config.MaxWorkers
	// maxMemoryMB := pipeline.Config.MaxMemoryMB

	inputQueue := make(chan fgl_metadata.InputMetadata[T])

	// Spawn fixed number of goroutines
	for i := 0; i < maxWorkers; i++ {
		go pipeline.worker(ctx, inputQueue)
	}

	// Send input values to the input channel
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
func convertCurrentOutToNextIn[T any](currentOut chan fgl_metadata.OutputMetadata[T]) chan fgl_metadata.InputMetadata[T] {
	nextIn := make(chan fgl_metadata.InputMetadata[T], 1)

	go func() {
		defer close(nextIn)

		for outMeta := range currentOut {
			nextIn <- fgl_metadata.InputMetadata[T]{
				InitialInput: &outMeta.InitialInput,
				Value:        outMeta.OutputValue,
				InputTime:    outMeta.InitialInput.InputTime,
			}
		}
	}()

	// Return the next input channel
	return nextIn
}

// processStageTransform goroutine
// It reads from the input channel, processes data, and sends results to the output channel.
// It returns an error if the stage fails.
func processStageTransform[T any](ctx context.Context, stage fgl_stage.Stage[T], in chan fgl_metadata.InputMetadata[T], out chan fgl_metadata.OutputMetadata[T], wg *sync.WaitGroup) error {
	go func() {
		defer func() {
			wg.Done()
			close(out)
		}()

		err := stage.TransformFunction(ctx, in, out)
		if err != nil {
			fmt.Printf("ERR: Failed to process stage %s: %s\n", stage.GetID(), err.Error())
			return
		}
	}()

	return nil
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
		if stageItem.Id == stage.Id {
			pipeline.Stages = append(pipeline.Stages[:index], pipeline.Stages[index+1:]...)
			break
		}
	}
}
