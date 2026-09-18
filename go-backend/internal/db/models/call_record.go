package models

import "time"

type CallRecord struct {
	ID              string    `json:"id"`
	PropertyID      string    `json:"property_id"`
	Transcript      string    `json:"transcript"`
	AISummary       string    `json:"ai_summary"`
	RecordingURL    string    `json:"recording_url"`
	DurationSeconds int       `json:"duration_seconds"`
	Status          string    `json:"status"` // success, failed, timeout
	CreatedAt       time.Time `json:"created_at"`
}

type CreateCallRecordRequest struct {
	PropertyID      string `json:"property_id" binding:"required"`
	Transcript      string `json:"transcript"`
	AISummary       string `json:"ai_summary"`
	RecordingURL    string `json:"recording_url"`
	DurationSeconds int    `json:"duration_seconds"`
	Status          string `json:"status"`
}
