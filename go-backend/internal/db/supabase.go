package db

import (
	"context"
	"encoding/json"
	"fmt"

	supa "github.com/supabase-community/supabase-go"
	"github.com/zumen-kun/go-backend/internal/config"
	"github.com/zumen-kun/go-backend/internal/db/models"
)

type SupabaseClient struct {
	client *supa.Client
}

func NewSupabaseClient(cfg *config.Config) (*SupabaseClient, error) {
	if cfg.SupabaseURL == "" || cfg.SupabaseKey == "" {
		return &SupabaseClient{}, nil // Return empty client for development
	}

	client, err := supa.NewClient(cfg.SupabaseURL, cfg.SupabaseKey, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Supabase client: %w", err)
	}

	return &SupabaseClient{client: client}, nil
}

// Property operations
func (s *SupabaseClient) CreateProperty(ctx context.Context, prop *models.Property) error {
	if s.client == nil {
		return fmt.Errorf("Supabase client not initialized")
	}

	data, err := json.Marshal(prop)
	if err != nil {
		return err
	}

	_, _, err = s.client.From("properties").Insert(data, false, "", "", "").Execute()
	return err
}

func (s *SupabaseClient) GetProperty(ctx context.Context, id string) (*models.Property, error) {
	if s.client == nil {
		return nil, fmt.Errorf("Supabase client not initialized")
	}

	data, _, err := s.client.From("properties").
		Select("*", "", false).
		Eq("id", id).
		Single().
		Execute()

	if err != nil {
		return nil, err
	}

	var property models.Property
	if err := json.Unmarshal(data, &property); err != nil {
		return nil, err
	}

	return &property, nil
}

func (s *SupabaseClient) ListProperties(ctx context.Context) ([]models.Property, error) {
	if s.client == nil {
		return []models.Property{}, nil
	}

	data, _, err := s.client.From("properties").
		Select("*", "", false).
		Order("created_at", nil).
		Execute()

	if err != nil {
		return nil, err
	}

	var properties []models.Property
	if err := json.Unmarshal(data, &properties); err != nil {
		return nil, err
	}

	return properties, nil
}

func (s *SupabaseClient) UpdateProperty(ctx context.Context, id string, updates map[string]interface{}) error {
	if s.client == nil {
		return fmt.Errorf("Supabase client not initialized")
	}

	data, err := json.Marshal(updates)
	if err != nil {
		return err
	}

	_, _, err = s.client.From("properties").
		Update(data, "", "").
		Eq("id", id).
		Execute()

	return err
}

func (s *SupabaseClient) DeleteProperty(ctx context.Context, id string) error {
	if s.client == nil {
		return fmt.Errorf("Supabase client not initialized")
	}

	_, _, err := s.client.From("properties").
		Delete("", "").
		Eq("id", id).
		Execute()

	return err
}

// CallRecord operations
func (s *SupabaseClient) CreateCallRecord(ctx context.Context, record *models.CallRecord) error {
	if s.client == nil {
		return fmt.Errorf("Supabase client not initialized")
	}

	data, err := json.Marshal(record)
	if err != nil {
		return err
	}

	_, _, err = s.client.From("call_records").Insert(data, false, "", "", "").Execute()
	return err
}

func (s *SupabaseClient) GetCallRecordsByProperty(ctx context.Context, propertyID string) ([]models.CallRecord, error) {
	if s.client == nil {
		return []models.CallRecord{}, nil
	}

	data, _, err := s.client.From("call_records").
		Select("*", "", false).
		Eq("property_id", propertyID).
		Order("created_at", nil).
		Execute()

	if err != nil {
		return nil, err
	}

	var records []models.CallRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, err
	}

	return records, nil
}
