package models

import "time"

type DockerfileAnalysis struct {
	ID           string                 `json:"id"`
	Filename     string                 `json:"filename"`
	Content      string                 `json:"content"`
	Instructions []DockerfileInstruction `json:"instructions"`
	Issues       []OptimizationIssue     `json:"issues"`
	Metrics      DockerfileMetrics       `json:"metrics"`
	CreatedAt    time.Time              `json:"created_at"`
}

type DockerfileInstruction struct {
	LineNumber int               `json:"line_number"`
	Command    string            `json:"command"`
	Arguments  []string          `json:"arguments"`
	Raw        string            `json:"raw"`
	Issues     []OptimizationIssue `json:"issues,omitempty"`
}

type OptimizationIssue struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	Message     string `json:"message"`
	LineNumber  int    `json:"line_number,omitempty"`
	Suggestion  string `json:"suggestion,omitempty"`
}

type DockerfileMetrics struct {
	TotalInstructions int     `json:"total_instructions"`
	LayerCount       int     `json:"layer_count"`
	EstimatedSize    string  `json:"estimated_size"`
	BuildTime        string  `json:"build_time,omitempty"`
	Efficiency       float64 `json:"efficiency"`
}