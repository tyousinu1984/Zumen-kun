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
	"github.com/zumen-kun/go-backend/internal/worker"
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

// ServeUploadedFile handles GET /uploads/*filepath - serves in-memory PDFs
func ServeUploadedFile(c *gin.Context) {
	filePath := c.Param("filepath")
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
	db  *db.SupabaseClient
	cfg *config.Config
}

func NewPDFHandler(db *db.SupabaseClient, cfg *config.Config) *PDFHandler {
	return &PDFHandler{
		db:  db,
		cfg: cfg,
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
		fileURL = fmt.Sprintf("http://go-backend:8080/uploads/%s", storageFilename)
	}

	// Create property record in Supabase
	property := &models.Property{
		ID:     propertyID,
		PDFURL: fileURL,
		Status: "parsing",
	}

	if dbErr := h.db.CreateProperty(c.Request.Context(), property); dbErr != nil {
		fmt.Printf("⚠️  DB CreateProperty failed: %v\n", dbErr)
	}

	// Launch Gemini parse goroutine (non-blocking)
	worker.StartParsePDF(worker.ParseJob{
		PropertyID: propertyID,
		PDFBytes:   fileBytes,
		GeminiKey:  h.cfg.GeminiAPIKey,
		DB:         h.db,
	})

	c.JSON(http.StatusCreated, gin.H{
		"property_id": propertyID,
		"pdf_url":     fileURL,
		"status":      "parsing",
		"message":     "PDF uploaded and parse task started",
	})
}

// GET /api/v1/properties/:id/pdf
// Serves the PDF for a given property ID from in-memory store or Supabase Storage
func (h *PDFHandler) ServePDF(c *gin.Context) {
	id := c.Param("id")

	// Look up property to get the pdf_url
	property, err := h.db.GetProperty(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Property not found"})
		return
	}

	if property.PDFURL == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "No PDF associated with this property"})
		return
	}

	// Extract storage key from URL (e.g. http://go-backend:8080/uploads/properties/xxx/...)
	const uploadsPrefix = "/uploads/"
	pdfURL := property.PDFURL
	if idx := strings.Index(pdfURL, uploadsPrefix); idx >= 0 {
		storageKey := pdfURL[idx+len(uploadsPrefix):]
		data, ok := globalPDFStore.Get(storageKey)
		if ok {
			c.Header("Content-Type", "application/pdf")
			c.Header("Content-Disposition", "inline; filename=\"zumen.pdf\"")
			c.Header("Access-Control-Allow-Origin", "*")
			c.Data(http.StatusOK, "application/pdf", data)
			return
		}
	}

	// Fallback: redirect to the stored URL (Supabase Storage or external)
	c.Redirect(http.StatusFound, pdfURL)
}

func (h *PDFHandler) uploadToStorage(c *gin.Context, filename string, data []byte) (string, error) {
	if h.cfg.SupabaseURL == "" {
		return "", fmt.Errorf("Supabase not configured")
	}

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

	publicURL := fmt.Sprintf("%s/storage/v1/object/public/property-pdfs/%s", h.cfg.SupabaseURL, filename)
	return publicURL, nil
}
