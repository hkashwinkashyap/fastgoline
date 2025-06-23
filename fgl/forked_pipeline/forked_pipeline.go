package fgl_forked_pipeline

import (
	"sync"
	"time"

	"github.com/hkashwinkashyap/fastgoline/fgl/metadata"
	fgl_stage "github.com/hkashwinkashyap/fastgoline/fgl/stage"
)

type ForkedOutputResult[T any] struct {
	InputID  string
	Output   T
	Duration time.Duration
	Err      error
}

type ForkedPipeline[T any] struct {
	Id          string
	Input       metadata.InputMetadata[T]
	Stages      []fgl_stage.Stage[T]
	Output      chan ForkedOutputResult[T]
	Done        chan struct{}
	StartedAt   time.Time
	CompletedAt time.Time
}

type ForkRegistry[T any] struct {
	Mu    sync.Mutex
	Forks map[string]*ForkedPipeline[T]
}
