package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zumen-kun/go-backend/internal/db"
)

type CallHandler struct {
	db *db.SupabaseClient
}

func NewCallHandler(db *db.SupabaseClient) *CallHandler {
	return &CallHandler{db: db}
}

// GET /api/v1/properties/:id/calls
func (h *CallHandler) GetCallsByProperty(c *gin.Context) {
	propertyID := c.Param("id")

	records, err := h.db.GetCallRecordsByProperty(c.Request.Context(), propertyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch call records"})
		return
	}

	c.JSON(http.StatusOK, records)
}
