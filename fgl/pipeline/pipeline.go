package fgl_pipeline

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hkashwinkashyap/fastgoline/fgl/metadata"
	fgl_stage "github.com/hkashwinkashyap/fastgoline/fgl/stage"

	"github.com/hkashwinkashyap/fastgoline/util"
)

// Pipeline holds a sequence of stages to process data.
// Each stage runs concurrently, passing data along channels.
type Pipeline[T any] struct {
	id            string
	Stages        []fgl_stage.Stage[T]
	InputChannel  <-chan metadata.InputMetadata[T]
	OutputChannel chan<- metadata.OutputMetadata[T]
	EnableForking bool
	ForkedInput   *metadata.InputMetadata[T]
	Done          chan struct{}
	StartedAt     time.Time
	CompletedAt   time.Time
	ForkedAt      *time.Time
	ForkRegistry  map[string]fgl_stage.Stage[T]
}

// GetID returns the ID of the pipeline.
func (pipeline *Pipeline[T]) GetID() string {
	return pipeline.id
}

// NewPipeline creates an empty pipeline instance.
func NewPipeline[T any](in <-chan metadata.InputMetadata[T], out chan<- metadata.OutputMetadata[T], enableForking bool) *Pipeline[T] {
	// Generate a unique id
	id := util.GenerateUUID()

	// If forking is enabled return a forked pipeline
	if enableForking {
		return &Pipeline[T]{
			id:            id,
			Stages:        nil,
			InputChannel:  in,
			OutputChannel: out,
			EnableForking: enableForking,
			ForkRegistry:  make(map[string]fgl_stage.Stage[T]),
			Done:          make(chan struct{}),
		}
	}

	// If forking is not enabled return a normal pipeline
	return &Pipeline[T]{id: id,
		Stages: nil, InputChannel: in,
		OutputChannel: out,
		EnableForking: enableForking,
		ForkRegistry:  nil,
	}
}

// RunPipeline runs the pipeline to process data.
// It returns an error if any stage fails.
func (pipeline *Pipeline[T]) RunPipeline(ctx context.Context) error {
	now := time.Now().UTC()
	fmt.Printf("TRACE: Starting pipeline %s at %s\n", pipeline.id, now)

	var err error
	var wg sync.WaitGroup

	// Set the number of goroutines to the number of stages
	wg.Add(len(pipeline.Stages))

	currentIn := pipeline.InputChannel

	for index, stage := range pipeline.Stages {
		currentOut := make(chan metadata.OutputMetadata[T])

		isFinal := index == len(pipeline.Stages)-1

		if isFinal {
			// Final stage
			// Pipe the final output to the provided output channel of the pipeline
			err = processStageTransform(ctx, stage, currentIn, pipeline.OutputChannel, &wg)
			if err != nil {
				ctx.Done()
				return err
			}
		} else {
			// Intermediate stage
			// Pipe the current output to the next input channel
			err = processStageTransform(ctx, stage, currentIn, currentOut, &wg)
			if err != nil {
				ctx.Done()
				return err
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

	fmt.Printf("TRACE: Finished Pipeline %s in %v ms\n", pipeline.GetID(), time.Now().UTC().Sub(now).Milliseconds())
	return nil
}

// convertCurrentOutToNextIn converts output metadata to input metadata
// This is used to pipe the output of a stage to the input of the next intermediate stage
func convertCurrentOutToNextIn[T any](currentOut <-chan metadata.OutputMetadata[T]) <-chan metadata.InputMetadata[T] {
	nextIn := make(chan metadata.InputMetadata[T])

	go func() {
		defer close(nextIn)

		for outMeta := range currentOut {
			nextIn <- metadata.InputMetadata[T]{
				Id:        outMeta.InputID,
				Value:     outMeta.Value,
				InputTime: outMeta.InputTime,
			}
		}
	}()

	// Return the next input channel
	return nextIn
}

// processStageTransform goroutine
// It reads from the input channel, processes data, and sends results to the output channel.
// It returns an error if any stage fails.
func processStageTransform[T any](ctx context.Context, stage fgl_stage.Stage[T], in <-chan metadata.InputMetadata[T], out chan<- metadata.OutputMetadata[T], wg *sync.WaitGroup) error {
	go func() {
		defer wg.Done()
		defer close(out)

		err := stage.TransformFunction(ctx, in, out)
		if err != nil {
			fmt.Printf("ERR: %s\n", err.Error())
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

// NewForkedPipeline creates a new forked pipeline instance
func NewForkedPipeline[T any](input metadata.InputMetadata[T], stages []fgl_stage.Stage[T], out chan<- metadata.OutputMetadata[T]) *Pipeline[T] {
	// Generate a unique id
	id := util.GenerateUUID()

	return &Pipeline[T]{
		id:            id,
		ForkedInput:   &input,
		Stages:        stages,
		Done:          make(chan struct{}),
		StartedAt:     time.Now().UTC(),
		OutputChannel: out,
	}
}

// RunForkedPipeline runs the forked pipeline
// It returns an error if any stage fails
// Pipes the output metadata to the provided metadata output channel
func (forkedPipeline *Pipeline[T]) RunForkedPipeline(ctx context.Context) {
	defer close(forkedPipeline.Done)

	value := forkedPipeline.ForkedInput.Value
	forkedAt := *forkedPipeline.ForkedAt

	fmt.Printf("TRACE: Forking pipeline %s at %s\n", forkedPipeline.id, forkedAt)

	var err error
	var wg sync.WaitGroup

	// Set the number of goroutines to the number of stages
	wg.Add(len(forkedPipeline.Stages))

	currentIn := make(<-chan metadata.InputMetadata[T])

outer:
	for index, stage := range forkedPipeline.Stages {
		currentOut := make(chan metadata.OutputMetadata[T])

		isFinal := index == len(forkedPipeline.Stages)-1

		select {
		case <-ctx.Done():
			fmt.Printf("ERR: Forked pipeline %s cancelled at %s - reason: %s\n", forkedPipeline.id, time.Now().UTC(), ctx.Err().Error())
			err = ctx.Err()
			break outer

		default:
			if isFinal {
				// Final stage
				// Pipe the final output to the provided output channel of the pipeline
				err = processStageTransform(ctx, stage, currentIn, forkedPipeline.OutputChannel, &wg)
				if err != nil {
					fmt.Printf("ERR: %s\n", err.Error())
					break
				}
			} else {
				// Intermediate stage
				// Pipe the output of the current stage to the input of the next stage
				err = processStageTransform(ctx, stage, currentIn, currentOut, &wg)
				if err != nil {
					fmt.Printf("ERR: %s\n", err.Error())
					break
				}
			}

			// Current output channel becomes the next input channel
			currentIn = convertCurrentOutToNextIn(currentOut)
		}
	}

	// Wait for all goroutines to finish
	wg.Wait()

	forkedPipeline.CompletedAt = time.Now().UTC()
	duration := forkedPipeline.CompletedAt.Sub(forkedAt).Milliseconds()
	totalDuration := forkedPipeline.CompletedAt.Sub(forkedPipeline.StartedAt).Milliseconds()

	fmt.Printf("TRACE: Completed forked pipeline %s in %v ms\n", forkedPipeline.id, duration)
	fmt.Printf("TRACE: Completed pipeline %s in %v ms\n", forkedPipeline.id, totalDuration)

	// Format outputMetadata
	outputMetadata := metadata.OutputMetadata[T]{
		InputID:    forkedPipeline.ForkedInput.Id,
		Value:      value,
		InputTime:  forkedPipeline.StartedAt,
		OutputTime: forkedPipeline.CompletedAt,
		ForkedAt:   &forkedAt,
		Duration:   duration,
		Err:        err,
	}

	forkedPipeline.Done <- struct{}{}

	forkedPipeline.OutputChannel <- outputMetadata
}

// GetForkedInput returns the input metadata of the forked pipeline
func (forkedPipeline *Pipeline[T]) GetForkedInput() metadata.InputMetadata[T] {
	return *forkedPipeline.ForkedInput
}

// GetStages returns the stages of the forked pipeline
func (forkedPipeline *Pipeline[T]) GetStages() []fgl_stage.Stage[T] {
	return forkedPipeline.Stages
}

// IsDone returns true if the forked pipeline is done
func (forkedPipeline *Pipeline[T]) IsDone() bool {
	select {
	case <-forkedPipeline.Done:
		return true

	default:
		return false
	}
}

// SetCompletedAt sets the completed at time of the forked pipeline
func (forkedPipeline *Pipeline[T]) SetCompletedAt(timestamp time.Time) {
	forkedPipeline.CompletedAt = timestamp
}

// MarkDone marks the forked pipeline as done
func (forkedPipeline *Pipeline[T]) MarkDone() {
	close(forkedPipeline.Done)
}
