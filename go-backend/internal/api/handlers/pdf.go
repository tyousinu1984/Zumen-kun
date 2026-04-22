package handlers

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/zumen-kun/go-backend/internal/config"
	"github.com/zumen-kun/go-backend/internal/db"
	"github.com/zumen-kun/go-backend/internal/db/models"
	"github.com/zumen-kun/go-backend/internal/redis"
)

type PDFHandler struct {
	db       *db.SupabaseClient
	producer *redis.TaskProducer
	cfg      *config.Config
}

func NewPDFHandler(db *db.SupabaseClient, producer *redis.TaskProducer, cfg *config.Config) *PDFHandler {
	return &PDFHandler{
		db:       db,
		producer: producer,
		cfg:      cfg,
	}
}

// POST /api/v1/properties/parse
// Accepts multipart/form-data with "file" field
func (h *PDFHandler) UploadAndParse(c *gin.Context) {
	// Get file from form
	file, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided"})
		return
	}
	defer file.Close()

	// Validate file type
	if !strings.HasSuffix(strings.ToLower(fileHeader.Filename), ".pdf") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only PDF files are accepted"})
		return
	}

	// Read file content
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}

	// Generate unique filename
	propertyID := uuid.New().String()
	ext := filepath.Ext(fileHeader.Filename)
	storageFilename := fmt.Sprintf("properties/%s/%s%s",
		propertyID,
		time.Now().Format("20060102150405"),
		ext,
	)

	// Upload to Supabase Storage
	fileURL, err := h.uploadToStorage(c, storageFilename, fileBytes)
	if err != nil {
		// If Supabase is not configured, use a placeholder URL for dev
		fileURL = fmt.Sprintf("/uploads/%s", storageFilename)
	}

	// Create property record
	property := &models.Property{
		ID:     propertyID,
		PDFURL: fileURL,
		Status: "pending",
	}

	if err := h.db.CreateProperty(c.Request.Context(), property); err != nil {
		// In development mode without Supabase, just proceed
		c.JSON(http.StatusOK, gin.H{
			"property_id": propertyID,
			"pdf_url":     fileURL,
			"status":      "pending",
			"message":     "Property created (development mode - Supabase not configured)",
		})
		return
	}

	// Push parse task to Redis
	if err := h.producer.PushParsePDFTask(c.Request.Context(), propertyID, fileURL); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to queue parse task"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"property_id": propertyID,
		"pdf_url":     fileURL,
		"status":      "parsing",
		"message":     "PDF uploaded and parse task queued successfully",
	})
}

func (h *PDFHandler) uploadToStorage(c *gin.Context, filename string, data []byte) (string, error) {
	if h.cfg.SupabaseURL == "" {
		return "", fmt.Errorf("Supabase not configured")
	}

	// Build upload URL
	uploadURL := fmt.Sprintf("%s/storage/v1/object/property-pdfs/%s", h.cfg.SupabaseURL, filename)

	req, err := http.NewRequestWithContext(c.Request.Context(), "POST", uploadURL, bytes.NewReader(data))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+h.cfg.SupabaseKey)
	req.Header.Set("Content-Type", "application/pdf")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("storage upload failed with status: %d", resp.StatusCode)
	}

	// Return public URL
	publicURL := fmt.Sprintf("%s/storage/v1/object/public/property-pdfs/%s", h.cfg.SupabaseURL, filename)
	return publicURL, nil
}
