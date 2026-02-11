package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

// InMemoryQueue implements Queue interface using Go channels
type InMemoryQueue struct {
	handlers  map[string]func(context.Context, []byte) error
	jobs      chan Job
	quit      chan struct{}
	wg        sync.WaitGroup
	mu        sync.RWMutex
	jobStatus map[string]string // Map to track job status
}

func NewInMemoryQueue(bufferSize int) *InMemoryQueue {
	if bufferSize <= 0 {
		bufferSize = 100
	}
	return &InMemoryQueue{
		handlers:  make(map[string]func(context.Context, []byte) error),
		jobs:      make(chan Job, bufferSize),
		quit:      make(chan struct{}),
		jobStatus: make(map[string]string),
	}
}

func (q *InMemoryQueue) Enqueue(ctx context.Context, jobType string, payload interface{}, opts ...Option) (string, error) {
	options := &Options{}
	for _, o := range opts {
		o(options)
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	jobID := uuid.New().String()
	job := Job{
		Type:    jobType,
		Payload: data,
		ID:      jobID,
	}

	q.mu.Lock()
	q.jobStatus[jobID] = "queued"
	q.mu.Unlock()

	// Handle delay if specified
	if options.Delay > 0 {
		go func() {
			select {
			case <-time.After(options.Delay):
				select {
				case q.jobs <- job:
				case <-q.quit:
				}
			case <-q.quit:
			}
		}()
		return jobID, nil
	}

	select {
	case q.jobs <- job:
		return jobID, nil
	case <-ctx.Done():
		return "", ctx.Err()
	case <-q.quit:
		return "", fmt.Errorf("queue is stopping")
	}
}

func (q *InMemoryQueue) RegisterHandler(jobType string, handler func(context.Context, []byte) error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.handlers[jobType] = handler
}

func (q *InMemoryQueue) Start() error {
	q.wg.Add(1)
	go q.worker()
	log.Println("In-memory queue started")
	return nil
}

func (q *InMemoryQueue) Stop() {
	close(q.quit)
	q.wg.Wait()
	log.Println("In-memory queue stopped")
}

func (q *InMemoryQueue) worker() {
	defer q.wg.Done()
	for {
		select {
		case job := <-q.jobs:
			q.processJob(job)
		case <-q.quit:
			return
		}
	}
}

func (q *InMemoryQueue) processJob(job Job) {
	q.mu.RLock()
	handler, ok := q.handlers[job.Type]
	q.mu.RUnlock()

	if !ok {
		log.Printf("No handler registered for job type: %s", job.Type)
		return
	}

	q.mu.Lock()
	q.jobStatus[job.ID] = "processing"
	q.mu.Unlock()

	// Create a context for the handler
	ctx := context.Background()
	err := handler(ctx, job.Payload)

	q.mu.Lock()
	if err != nil {
		log.Printf("Error processing job %s (Type: %s): %v", job.ID, job.Type, err)
		q.jobStatus[job.ID] = "failed"
	} else {
		log.Printf("Successfully processed job %s (Type: %s)", job.ID, job.Type)
		q.jobStatus[job.ID] = "completed"
	}
	q.mu.Unlock()
}

func (q *InMemoryQueue) GetJobStatus(ctx context.Context, jobID string) (string, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	status, ok := q.jobStatus[jobID]
	if !ok {
		return "", fmt.Errorf("job not found")
	}
	return status, nil
}
