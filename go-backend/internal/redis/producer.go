package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

const (
	QueueParsePDF  = "task:parse_pdf"
	QueueVoiceCall = "task:voice_call"
)

type TaskProducer struct {
	client *Client
}

func NewTaskProducer(client *Client) *TaskProducer {
	return &TaskProducer{client: client}
}

type ParsePDFTask struct {
	PropertyID string `json:"property_id"`
	FileURL    string `json:"file_url"`
	Timestamp  int64  `json:"timestamp"`
}

type VoiceCallTask struct {
	PropertyID  string `json:"property_id"`
	PhoneNumber string `json:"phone_number"`
	PropertyName string `json:"property_name"`
	RoomNumber  string `json:"room_number"`
	Timestamp   int64  `json:"timestamp"`
}

func (p *TaskProducer) PushParsePDFTask(ctx context.Context, propertyID, fileURL string) error {
	task := ParsePDFTask{
		PropertyID: propertyID,
		FileURL:    fileURL,
		Timestamp:  time.Now().Unix(),
	}

	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	if err := p.client.rdb.RPush(ctx, QueueParsePDF, data).Err(); err != nil {
		return fmt.Errorf("failed to push task to Redis: %w", err)
	}

	return nil
}

func (p *TaskProducer) PushVoiceCallTask(ctx context.Context, propertyID, phoneNumber, propertyName, roomNumber string) error {
	task := VoiceCallTask{
		PropertyID:   propertyID,
		PhoneNumber:  phoneNumber,
		PropertyName: propertyName,
		RoomNumber:   roomNumber,
		Timestamp:    time.Now().Unix(),
	}

	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	if err := p.client.rdb.RPush(ctx, QueueVoiceCall, data).Err(); err != nil {
		return fmt.Errorf("failed to push task to Redis: %w", err)
	}

	return nil
}
