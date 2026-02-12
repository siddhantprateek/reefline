package models

import (
	"time"

	"gorm.io/gorm"
)

// JobStatus represents the state of a job
type JobStatus string

const (
	JobStatusPending   JobStatus = "PENDING"
	JobStatusQueued    JobStatus = "QUEUED"
	JobStatusRunning   JobStatus = "RUNNING"
	JobStatusCompleted JobStatus = "COMPLETED"
	JobStatusFailed    JobStatus = "FAILED"
	JobStatusCancelled JobStatus = "CANCELLED"
	JobStatusSkipped   JobStatus = "SKIPPED"
	JobStatusUnknown   JobStatus = "UNKNOWN"
)

// Job represents an analysis task
type Job struct {
	ID           string         `json:"id" gorm:"primaryKey"`
	JobID        string         `json:"job_id" gorm:"uniqueIndex"`
	UserID       string         `json:"user_id" gorm:"index"`
	ImageRef     string         `json:"image_ref"`
	Dockerfile   string         `json:"dockerfile" gorm:"type:text"`
	Status       JobStatus      `json:"status" gorm:"index"`
	Scenario     string         `json:"scenario"`                  // "dockerfile", "image", "both"
	Metadata     string         `json:"metadata" gorm:"type:text"` // JSON string of Skopeo results, etc.
	ErrorMessage string         `json:"error_message" gorm:"type:text"`
	Progress     int            `json:"progress"` // 0-100
	QueuedAt     *time.Time     `json:"queued_at"`
	StartedAt    *time.Time     `json:"started_at" gorm:"index:idx_timing"`
	CompletedAt  *time.Time     `json:"completed_at"`
	ToolMetrics  string         `json:"tool_metrics" gorm:"type:text"` // JSON string of per-tool timing data
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// BeforeCreate hooks into GORM to set UUID if needed
func (j *Job) BeforeCreate(tx *gorm.DB) (err error) {
	// Let uuid generation happen in handler for now or add lib
	return
}

// GetQueueWaitDuration returns the time spent waiting in queue before processing started
// Returns 0 if QueuedAt or StartedAt is nil
func (j *Job) GetQueueWaitDuration() time.Duration {
	if j.QueuedAt == nil || j.StartedAt == nil {
		return 0
	}
	return j.StartedAt.Sub(*j.QueuedAt)
}

// GetProcessingDuration returns the time spent processing the job
// Returns 0 if StartedAt or CompletedAt is nil
func (j *Job) GetProcessingDuration() time.Duration {
	if j.StartedAt == nil || j.CompletedAt == nil {
		return 0
	}
	return j.CompletedAt.Sub(*j.StartedAt)
}

// GetTotalDuration returns the total time from queue to completion
// Returns 0 if QueuedAt or CompletedAt is nil
func (j *Job) GetTotalDuration() time.Duration {
	if j.QueuedAt == nil || j.CompletedAt == nil {
		return 0
	}
	return j.CompletedAt.Sub(*j.QueuedAt)
}
