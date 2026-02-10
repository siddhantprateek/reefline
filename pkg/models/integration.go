package models

import "time"

// Integration represents a user's connection to an external service
// (GitHub, Docker Hub, Harbor, AI providers).
// Credentials are stored encrypted at rest.
type Integration struct {
	ID            uint       `json:"id" gorm:"primaryKey"`
	UserID        string     `json:"user_id" gorm:"index;not null"`
	IntegrationID string     `json:"integration_id" gorm:"not null"`      // e.g. "github", "docker", "harbor", "openai"
	Status        string     `json:"status" gorm:"default:disconnected"`  // "connected", "disconnected", "error"
	Credentials   string     `json:"-" gorm:"type:text"`                  // encrypted JSON blob — never exposed in API responses
	Metadata      string     `json:"metadata,omitempty" gorm:"type:text"` // JSON — public info like username, provider version
	ConnectedAt   *time.Time `json:"connected_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// TableName overrides the default GORM table name
func (Integration) TableName() string {
	return "integrations"
}
