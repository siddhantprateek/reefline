package models

import (
	"time"

	"gorm.io/gorm"
)

// JobStatus represents the state of a job
type JobStatus string

const (
	JobStatusQueued    JobStatus = "QUEUED"
	JobStatusRunning   JobStatus = "RUNNING"
	JobStatusCompleted JobStatus = "COMPLETED"
	JobStatusFailed    JobStatus = "FAILED"
)

// Job represents an analysis task
type Job struct {
	ID        string         `json:"id" gorm:"primaryKey"`
	UserID    string         `json:"user_id" gorm:"index"`
	Status    JobStatus      `json:"status" gorm:"index"`
	Scenario  string         `json:"scenario"` // "dockerfile_only", "image_only", "both"
	Metadata  string         `json:"metadata"` // JSON string of Skopeo results, etc.
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// BeforeCreate hooks into GORM to set UUID if needed
func (j *Job) BeforeCreate(tx *gorm.DB) (err error) {
	// Let uuid generation happen in handler for now or add lib
	return
}
