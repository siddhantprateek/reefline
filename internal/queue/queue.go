package queue

import (
	"context"
	"encoding/json"
	"time"
)

// Job represents a task to be processed
type Job struct {
	Type    string
	Payload []byte
	ID      string
}

// Queue defines the interface for job queue operations
type Queue interface {
	// Enqueue adds a job to the queue
	Enqueue(ctx context.Context, jobType string, payload interface{}, opts ...Option) (string, error)

	// RegisterHandler registers a handler for a specific job type
	RegisterHandler(jobType string, handler func(ctx context.Context, payload []byte) error)

	// Start starts the queue workers
	Start() error

	// Stop stops the queue workers
	Stop()

	// GetJobStatus returns the status of a job
	GetJobStatus(ctx context.Context, jobID string) (string, error)

	// Stats returns queue statistics
	Stats(ctx context.Context) (*QueueStats, error)
}

// QueueStats represents queue statistics
type QueueStats struct {
	Active    int `json:"active"`
	Pending   int `json:"pending"`
	Scheduled int `json:"scheduled"`
	Completed int `json:"completed"`
	Failed    int `json:"failed"`
}

// Option represents queue options (e.g., delay, priority)
type Option func(*Options)

type Options struct {
	Delay time.Duration
}

func WithDelay(d time.Duration) Option {
	return func(o *Options) {
		o.Delay = d
	}
}

// UnmarshalPayload is a helper to unmarshal job payload
func UnmarshalPayload(payload []byte, v interface{}) error {
	return json.Unmarshal(payload, v)
}
