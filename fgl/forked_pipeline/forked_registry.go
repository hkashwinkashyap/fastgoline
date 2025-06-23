package fgl_forked_pipeline

import "sync"

// ForkRegistry is a map of forked pipelines
// It is used to store the state of all forked pipelines
type ForkRegistry[T any] struct {
	Mutex           sync.Mutex
	ForkedPipelines map[string]*ForkedPipeline[T]
}

// NewForkRegistry creates a new fork registry
func NewForkRegistry[T any]() *ForkRegistry[T] {
	return &ForkRegistry[T]{
		ForkedPipelines: make(map[string]*ForkedPipeline[T]),
	}
}

// Thread-safe Add adds a forked pipeline to the registry
func (forkedRegistry *ForkRegistry[T]) Add(forkedPipeline *ForkedPipeline[T]) {
	forkedRegistry.Mutex.Lock()
	defer forkedRegistry.Mutex.Unlock()
	forkedRegistry.ForkedPipelines[forkedPipeline.Id] = forkedPipeline
}

// Thread-safe Get returns a forked pipeline from the registry
func (forkedRegistry *ForkRegistry[T]) Get(id string) (*ForkedPipeline[T], bool) {
	forkedRegistry.Mutex.Lock()
	defer forkedRegistry.Mutex.Unlock()
	forkedPipeline, exists := forkedRegistry.ForkedPipelines[id]
	return forkedPipeline, exists
}

// Thread-safe Remove removes a forked pipeline from the registry
func (forkedRegistry *ForkRegistry[T]) Remove(id string) {
	forkedRegistry.Mutex.Lock()
	defer forkedRegistry.Mutex.Unlock()
	delete(forkedRegistry.ForkedPipelines, id)
}
