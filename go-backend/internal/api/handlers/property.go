package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/zumen-kun/go-backend/internal/db"
	"github.com/zumen-kun/go-backend/internal/db/models"
	"github.com/zumen-kun/go-backend/internal/redis"
)

type PropertyHandler struct {
	db       *db.SupabaseClient
	producer *redis.TaskProducer
}

func NewPropertyHandler(db *db.SupabaseClient, producer *redis.TaskProducer) *PropertyHandler {
	return &PropertyHandler{
		db:       db,
		producer: producer,
	}
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

// POST /api/v1/properties/:id/parse
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

	// Update status to parsing
	if err := h.db.UpdateProperty(c.Request.Context(), id, map[string]interface{}{"status": "parsing"}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update status"})
		return
	}

	// Push task to Redis
	if err := h.producer.PushParsePDFTask(c.Request.Context(), id, property.PDFURL); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to queue parse task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Parse task queued successfully"})
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

	// Push task to Redis
	if err := h.producer.PushVoiceCallTask(c.Request.Context(), id, property.PhoneNumber, property.Name, property.RoomNumber); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to queue call task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Voice call task queued successfully"})
}
