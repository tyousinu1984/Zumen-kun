package handlers

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/zumen-kun/go-backend/internal/config"
	"github.com/zumen-kun/go-backend/internal/db"
	"github.com/zumen-kun/go-backend/internal/db/models"
	"github.com/zumen-kun/go-backend/internal/redis"
)

// ─────────────────────────────────────────────────────────
// In-memory PDF store (開発環境用: Supabase Storage の代替)
// ─────────────────────────────────────────────────────────

type pdfStore struct {
	mu    sync.RWMutex
	files map[string][]byte // key: "properties/{propertyID}/{filename}"
}

var globalPDFStore = &pdfStore{
	files: make(map[string][]byte),
}

func (s *pdfStore) Save(key string, data []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.files[key] = data
}

func (s *pdfStore) Get(key string) ([]byte, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	data, ok := s.files[key]
	return data, ok
}

// ServeFile handles GET /uploads/*filepath - serves in-memory PDFs
func ServeUploadedFile(c *gin.Context) {
	filePath := c.Param("filepath")
	// Remove leading slash
	filePath = strings.TrimPrefix(filePath, "/")

	data, ok := globalPDFStore.Get(filePath)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "inline")
	c.Data(http.StatusOK, "application/pdf", data)
}

// ─────────────────────────────────────────────────────────
// PDF Handler
// ─────────────────────────────────────────────────────────

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

	// Generate unique property ID and filename
	propertyID := uuid.New().String()
	ext := filepath.Ext(fileHeader.Filename)
	storageFilename := fmt.Sprintf("properties/%s/%s%s",
		propertyID,
		time.Now().Format("20060102150405"),
		ext,
	)

	// Try to upload to Supabase Storage first
	fileURL, uploadErr := h.uploadToStorage(c, storageFilename, fileBytes)
	if uploadErr != nil {
		// Fallback: save to in-memory store and serve locally
		globalPDFStore.Save(storageFilename, fileBytes)
		// Use the go-backend internal URL so python-worker can fetch it via Docker network
		fileURL = fmt.Sprintf("http://go-backend:8080/uploads/%s", storageFilename)
	}

	// Create property record in Supabase
	property := &models.Property{
		ID:     propertyID,
		PDFURL: fileURL,
		Status: "parsing",
	}

	dbErr := h.db.CreateProperty(c.Request.Context(), property)
	if dbErr != nil {
		// Log but continue - Redis task is more important in dev mode
		fmt.Printf("⚠️  DB CreateProperty failed (dev mode): %v\n", dbErr)
	}

	// Push parse task to Redis (ALWAYS - regardless of DB/Storage errors)
	if h.producer == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Redis not connected - cannot queue parse task"})
		return
	}

	if err := h.producer.PushParsePDFTask(c.Request.Context(), propertyID, fileURL); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to queue parse task: %v", err)})
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
