package fgl_pipeline

import "sync"

// ForkRegistry is a map of forked pipelines
// It is used to store the state of all forked pipelines
type ForkRegistry[T any] struct {
	Mutex           sync.Mutex
	ForkedPipelines map[string]*Pipeline[T]
}

// NewForkRegistry creates a new fork registry
func NewForkRegistry[T any]() *ForkRegistry[T] {
	return &ForkRegistry[T]{
		ForkedPipelines: make(map[string]*Pipeline[T]),
	}
}

// Thread-safe Add adds a forked pipeline to the registry
func (forkedRegistry *ForkRegistry[T]) Add(forkedPipeline *Pipeline[T]) {
	forkedRegistry.Mutex.Lock()
	defer forkedRegistry.Mutex.Unlock()

	// Get the id of the forked pipeline
	id := forkedPipeline.GetID()
	forkedRegistry.ForkedPipelines[id] = forkedPipeline
}

// Thread-safe Get returns a forked pipeline from the registry
func (forkedRegistry *ForkRegistry[T]) Get(id string) (*Pipeline[T], bool) {
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
