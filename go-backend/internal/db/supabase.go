package db

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	supa "github.com/supabase-community/supabase-go"
	"github.com/zumen-kun/go-backend/internal/config"
	"github.com/zumen-kun/go-backend/internal/db/models"
)

type SupabaseClient struct {
	client     *supa.Client
	restURL    string
	serviceKey string
	httpClient *http.Client
}

func NewSupabaseClient(cfg *config.Config) (*SupabaseClient, error) {
	if cfg.SupabaseURL == "" || cfg.SupabaseKey == "" {
		return &SupabaseClient{}, nil
	}

	client, err := supa.NewClient(cfg.SupabaseURL, cfg.SupabaseKey, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Supabase client: %w", err)
	}

	return &SupabaseClient{
		client:     client,
		restURL:    cfg.SupabaseURL + "/rest/v1",
		serviceKey: cfg.SupabaseKey,
		httpClient: &http.Client{},
	}, nil
}

// postREST sends a POST request directly to PostgREST
func (s *SupabaseClient) postREST(ctx context.Context, table string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/%s", s.restURL, table)
	log.Printf("🗃️  POST %s payload: %s", url, string(data))

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", s.serviceKey)
	req.Header.Set("Authorization", "Bearer "+s.serviceKey)
	req.Header.Set("Prefer", "return=minimal")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP POST failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	log.Printf("📥 POST response %d: %s", resp.StatusCode, string(body))

	if resp.StatusCode >= 300 {
		return fmt.Errorf("PostgREST error %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// patchREST sends a PATCH request directly to PostgREST
func (s *SupabaseClient) patchREST(ctx context.Context, table string, id string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/%s?id=eq.%s", s.restURL, table, id)
	req, err := http.NewRequestWithContext(ctx, "PATCH", url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", s.serviceKey)
	req.Header.Set("Authorization", "Bearer "+s.serviceKey)
	req.Header.Set("Prefer", "return=minimal")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP PATCH failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return fmt.Errorf("PostgREST PATCH error %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// Property operations
func (s *SupabaseClient) CreateProperty(ctx context.Context, prop *models.Property) error {
	if s.restURL == "" {
		return fmt.Errorf("Supabase client not initialized")
	}
	return s.postREST(ctx, "properties", prop)
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
	if s.restURL == "" {
		return []models.Property{}, nil
	}

	url := fmt.Sprintf("%s/properties?order=page_number.asc.nullslast,created_at.asc", s.restURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("apikey", s.serviceKey)
	req.Header.Set("Authorization", "Bearer "+s.serviceKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var properties []models.Property
	if err := json.Unmarshal(body, &properties); err != nil {
		return nil, err
	}
	return properties, nil
}

func (s *SupabaseClient) UpdateProperty(ctx context.Context, id string, updates map[string]interface{}) error {
	if s.restURL == "" {
		return fmt.Errorf("Supabase client not initialized")
	}
	return s.patchREST(ctx, "properties", id, updates)
}

func (s *SupabaseClient) DeleteProperty(ctx context.Context, id string) error {
	if s.restURL == "" {
		return fmt.Errorf("Supabase client not initialized")
	}
	url := fmt.Sprintf("%s/properties?id=eq.%s", s.restURL, id)
	return s.deleteREST(ctx, url)
}

func (s *SupabaseClient) DeleteAllProperties(ctx context.Context) error {
	if s.restURL == "" {
		return fmt.Errorf("Supabase client not initialized")
	}
	// id=neq.00000000... は全件にマッチするトリック
	url := fmt.Sprintf("%s/properties?id=neq.00000000-0000-0000-0000-000000000000", s.restURL)
	return s.deleteREST(ctx, url)
}

func (s *SupabaseClient) DeleteBulkProperties(ctx context.Context, ids []string) error {
	if s.restURL == "" {
		return fmt.Errorf("Supabase client not initialized")
	}
	// PostgREST: ?id=in.(id1,id2,...)
	inClause := "("
	for i, id := range ids {
		if i > 0 {
			inClause += ","
		}
		inClause += id
	}
	inClause += ")"
	url := fmt.Sprintf("%s/properties?id=in.%s", s.restURL, inClause)
	return s.deleteREST(ctx, url)
}

func (s *SupabaseClient) deleteREST(ctx context.Context, url string) error {
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("apikey", s.serviceKey)
	req.Header.Set("Authorization", "Bearer "+s.serviceKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP DELETE failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return fmt.Errorf("PostgREST DELETE error %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// CallRecord operations
func (s *SupabaseClient) CreateCallRecord(ctx context.Context, record *models.CallRecord) error {
	if s.restURL == "" {
		return fmt.Errorf("Supabase client not initialized")
	}
	return s.postREST(ctx, "call_records", record)
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

// ============================================================
// Company operations
// ============================================================

func (s *SupabaseClient) CreateCompany(ctx context.Context, id, name string) error {
	if s.restURL == "" {
		return fmt.Errorf("Supabase client not initialized")
	}
	payload := map[string]interface{}{
		"id":   id,
		"name": name,
	}
	return s.postREST(ctx, "companies", payload)
}

// ============================================================
// User operations
// ============================================================

func (s *SupabaseClient) CreateUser(ctx context.Context, user *models.User) error {
	if s.restURL == "" {
		return fmt.Errorf("Supabase client not initialized")
	}
	payload := map[string]interface{}{
		"id":            user.ID,
		"company_id":    user.CompanyID,
		"email":         user.Email,
		"password_hash": user.PasswordHash,
		"role":          user.Role,
		"display_name":  user.DisplayName,
		"is_active":     user.IsActive,
	}
	return s.postREST(ctx, "users", payload)
}

func (s *SupabaseClient) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	if s.restURL == "" {
		return nil, fmt.Errorf("Supabase client not initialized")
	}

	url := fmt.Sprintf("%s/users?email=eq.%s&limit=1", s.restURL, email)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("apikey", s.serviceKey)
	req.Header.Set("Authorization", "Bearer "+s.serviceKey)
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var users []models.User
	if err := json.Unmarshal(body, &users); err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, fmt.Errorf("user not found")
	}
	return &users[0], nil
}
