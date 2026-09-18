package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/zumen-kun/go-backend/internal/config"
	"github.com/zumen-kun/go-backend/internal/db"
	"github.com/zumen-kun/go-backend/internal/db/models"
	"github.com/zumen-kun/go-backend/internal/worker"
)

type PropertyHandler struct {
	db  *db.SupabaseClient
	cfg *config.Config
}

func NewPropertyHandler(db *db.SupabaseClient, cfg *config.Config) *PropertyHandler {
	return &PropertyHandler{db: db, cfg: cfg}
}

// POST /api/v1/properties
func (h *PropertyHandler) CreateProperty(c *gin.Context) {
	var req models.CreatePropertyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	property := &models.Property{
		ID:          uuid.New().String(),
		Name:        req.Name,
		RoomNumber:  req.RoomNumber,
		PhoneNumber: req.PhoneNumber,
		Price:       req.Price,
		AdRate:      req.AdRate,
		PDFURL:      req.PDFURL,
		Status:      "pending",
	}

	if err := h.db.CreateProperty(c.Request.Context(), property); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create property"})
		return
	}

	c.JSON(http.StatusCreated, property)
}

// GET /api/v1/properties
func (h *PropertyHandler) ListProperties(c *gin.Context) {
	properties, err := h.db.ListProperties(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch properties"})
		return
	}

	c.JSON(http.StatusOK, properties)
}

// GET /api/v1/properties/:id
func (h *PropertyHandler) GetProperty(c *gin.Context) {
	id := c.Param("id")

	property, err := h.db.GetProperty(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Property not found"})
		return
	}

	c.JSON(http.StatusOK, property)
}

// PATCH /api/v1/properties/:id
func (h *PropertyHandler) UpdateProperty(c *gin.Context) {
	id := c.Param("id")

	var req models.UpdatePropertyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.RoomNumber != nil {
		updates["room_number"] = *req.RoomNumber
	}
	if req.PhoneNumber != nil {
		updates["phone_number"] = *req.PhoneNumber
	}
	if req.Price != nil {
		updates["price"] = *req.Price
	}
	if req.AdRate != nil {
		updates["ad_rate"] = *req.AdRate
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.ForeignerOK != nil {
		updates["foreigner_ok"] = *req.ForeignerOK
	}

	if err := h.db.UpdateProperty(c.Request.Context(), id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update property"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Property updated successfully"})
}

// DELETE /api/v1/properties/:id
func (h *PropertyHandler) DeleteProperty(c *gin.Context) {
	id := c.Param("id")

	if err := h.db.DeleteProperty(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete property"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Property deleted successfully"})
}

// DELETE /api/v1/properties  (全削除)
func (h *PropertyHandler) DeleteAllProperties(c *gin.Context) {
	if err := h.db.DeleteAllProperties(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete all properties"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "All properties deleted"})
}

// DELETE /api/v1/properties/bulk  (複数削除)
type BulkDeleteRequest struct {
	IDs []string `json:"ids"`
}

func (h *PropertyHandler) BulkDeleteProperties(c *gin.Context) {
	var req BulkDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil || len(req.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ids array is required"})
		return
	}

	if err := h.db.DeleteBulkProperties(c.Request.Context(), req.IDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete properties"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Properties deleted", "count": len(req.IDs)})
}

// POST /api/v1/properties/:id/parse
// Re-triggers parsing for an existing property using in-memory PDF store
func (h *PropertyHandler) TriggerParse(c *gin.Context) {
	id := c.Param("id")

	property, err := h.db.GetProperty(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Property not found"})
		return
	}

	if property.PDFURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "PDF URL not found"})
		return
	}

	// in-memory store からPDFバイトを取得
	const uploadsPrefix = "/uploads/"
	pdfBytes := []byte(nil)
	if idx := strings.Index(property.PDFURL, uploadsPrefix); idx >= 0 {
		storageKey := property.PDFURL[idx+len(uploadsPrefix):]
		if data, ok := globalPDFStore.Get(storageKey); ok {
			pdfBytes = data
		}
	}

	if len(pdfBytes) == 0 {
		// PDFが見つからない場合はエラー（コンテナ再起動でメモリが消える）
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "PDF data not found in memory. Please re-upload the PDF file.",
		})
		return
	}

	// Update status to parsing
	if err := h.db.UpdateProperty(c.Request.Context(), id, map[string]interface{}{"status": "parsing"}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update status"})
		return
	}

	// Launch Gemini parse goroutine (non-blocking)
	worker.StartParsePDF(worker.ParseJob{
		PropertyID: id,
		PDFBytes:   pdfBytes,
		GeminiKey:  h.cfg.GeminiAPIKey,
		DB:         h.db,
	})

	c.JSON(http.StatusOK, gin.H{
		"message":     "AI re-parse started",
		"pdf_url":     property.PDFURL,
		"property_id": id,
	})
}

// POST /api/v1/properties/:id/call
func (h *PropertyHandler) TriggerCall(c *gin.Context) {
	id := c.Param("id")

	property, err := h.db.GetProperty(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Property not found"})
		return
	}

	if property.PhoneNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Phone number not found"})
		return
	}

	// Update status to calling
	if err := h.db.UpdateProperty(c.Request.Context(), id, map[string]interface{}{"status": "calling"}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "AI call triggered (Retell AI integration pending)",
		"property_id":   id,
		"phone_number":  property.PhoneNumber,
		"property_name": property.Name,
	})
}

// POST /api/v1/properties/:id/cancel-call
func (h *PropertyHandler) CancelCall(c *gin.Context) {
	id := c.Param("id")

	_, err := h.db.GetProperty(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Property not found"})
		return
	}

	if err := h.db.UpdateProperty(c.Request.Context(), id, map[string]interface{}{"status": "pending"}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel call"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Call cancelled",
		"property_id": id,
	})
}
