package worker

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/zumen-kun/go-backend/internal/db"
	"github.com/zumen-kun/go-backend/internal/db/models"
)

// ParseJob represents a PDF parse job
type ParseJob struct {
	PropertyID string
	PDFBytes   []byte
	GeminiKey  string
	DB         *db.SupabaseClient
}

// StartParsePDF launches a goroutine to parse a PDF with Gemini.
// It updates the first property record (propertyID) and creates
// additional records for extra properties found in the PDF.
func StartParsePDF(job ParseJob) {
	go func() {
		start := time.Now()
		log.Printf("🔍 [worker] Starting PDF parse for property: %s", job.PropertyID)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		// Call Gemini with PDF bytes directly
		properties, err := ParsePDFWithGemini(ctx, job.PDFBytes, job.GeminiKey)
		if err != nil {
			log.Printf("❌ [worker] Gemini parse failed for %s: %v", job.PropertyID, err)
			// Mark as failed in DB
			_ = job.DB.UpdateProperty(ctx, job.PropertyID, map[string]interface{}{
				"status": "failed",
			})
			return
		}

		if len(properties) == 0 {
			log.Printf("⚠️  [worker] No properties extracted from PDF: %s", job.PropertyID)
			_ = job.DB.UpdateProperty(ctx, job.PropertyID, map[string]interface{}{
				"status": "failed",
			})
			return
		}

		// 元PDFのURL（全物件で共有）— job.PropertyIDのpdf_urlを参照
		sourceProp, dbErr := job.DB.GetProperty(ctx, job.PropertyID)
		sourcePDFURL := ""
		if dbErr == nil {
			sourcePDFURL = sourceProp.PDFURL
		}

		// Update the first property record
		first := properties[0]
		updates := buildPropertyUpdates(first)
		updates["pdf_url"] = sourcePDFURL
		updates["source_pdf_id"] = job.PropertyID
		// Gemini が返した page_number を使用（0の場合は1にフォールバック）
		firstPage := first.PageNumber
		if firstPage <= 0 {
			firstPage = 1
		}
		updates["page_number"] = firstPage
		if err := job.DB.UpdateProperty(ctx, job.PropertyID, updates); err != nil {
			log.Printf("⚠️  [worker] Failed to update property %s: %v", job.PropertyID, err)
		} else {
			log.Printf("✅ [worker] Updated property %s (page 1): %s", job.PropertyID, first.PropertyName)
		}

		// Create additional property records for pages 2+
		for i := 1; i < len(properties); i++ {
			p := properties[i]
			// Gemini が返した page_number を使用（0の場合は i+1 にフォールバック）
			pageNum := p.PageNumber
			if pageNum <= 0 {
				pageNum = i + 1
			}
			newProp := &models.Property{
				ID:          uuid.New().String(),
				Name:        p.PropertyName,
				Status:      "completed",
				PDFURL:      sourcePDFURL,   // 同じPDFを参照
				SourcePDFID: job.PropertyID, // 元PDFのpropertyID
				PageNumber:  pageNum,        // Gemini が判定した実際のページ番号
			}
			if p.RoomNumber != "" {
				newProp.RoomNumber = p.RoomNumber
			}
			if p.PhoneNumber != "" {
				newProp.PhoneNumber = p.PhoneNumber
			}
			if p.Price != nil {
				newProp.Price = *p.Price
			}
			if p.AdRate != "" {
				newProp.AdRate = p.AdRate
			}

			if err := job.DB.CreateProperty(ctx, newProp); err != nil {
				log.Printf("⚠️  [worker] Failed to create property for page %d: %v", i+1, err)
			} else {
				log.Printf("✅ [worker] Created property for page %d: %s", i+1, p.PropertyName)
			}
		}

		log.Printf("🎉 [worker] PDF parse completed for %s (%d properties, %.1fs)",
			job.PropertyID, len(properties), time.Since(start).Seconds())
	}()
}

// buildPropertyUpdates converts ParsedProperty to a DB update map
func buildPropertyUpdates(p ParsedProperty) map[string]interface{} {
	updates := map[string]interface{}{
		"status": "completed",
	}
	if p.PropertyName != "" {
		updates["name"] = p.PropertyName
	}
	if p.RoomNumber != "" {
		updates["room_number"] = p.RoomNumber
	}
	if p.PhoneNumber != "" {
		updates["phone_number"] = p.PhoneNumber
	}
	if p.Price != nil {
		updates["price"] = fmt.Sprintf("%.1f", *p.Price)
	}
	if p.AdRate != "" {
		updates["ad_rate"] = p.AdRate
	}
	return updates
}
