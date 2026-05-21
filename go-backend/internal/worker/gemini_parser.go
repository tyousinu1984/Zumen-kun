package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"google.golang.org/genai"
)

// ParsedProperty holds the extracted property data from Gemini
type ParsedProperty struct {
	PropertyName string   `json:"property_name"`
	RoomNumber   string   `json:"room_number"`
	PhoneNumber  string   `json:"phone_number"`
	Price        *float64 `json:"price"`
	AdRate       string   `json:"ad_rate"`
	Address      string   `json:"address"`
	PageNumber   int      `json:"page_number"` // Gemini が判定した実際のページ番号
}

// ParsePDFWithGemini sends PDF bytes directly to Gemini 2.0 Flash
// and returns a list of extracted properties (one per page/unit).
func ParsePDFWithGemini(ctx context.Context, pdfBytes []byte, apiKey string) ([]ParsedProperty, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
		HTTPOptions: genai.HTTPOptions{
			APIVersion: GeminiAPIVersion,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	log.Printf("🤖 Sending PDF (%d bytes) to Gemini [%s]...", len(pdfBytes), GeminiModel)

	// Send PDF directly - no pdf2image conversion needed!
	contents := []*genai.Content{
		genai.NewContentFromParts(
			[]*genai.Part{
				genai.NewPartFromBytes(pdfBytes, PDFMimeType),
				genai.NewPartFromText(PropertyParsePrompt),
			},
			genai.RoleUser,
		),
	}

	resp, err := client.Models.GenerateContent(ctx, GeminiModel, contents, nil)
	if err != nil {
		return nil, fmt.Errorf("Gemini API error: %w", err)
	}

	if len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("Gemini returned no candidates")
	}

	content := resp.Text()
	log.Printf("📝 Gemini response (first 300 chars): %s", truncate(content, 300))

	return parseJSONResponse(content)
}

// parseJSONResponse extracts JSON array from Gemini response text
func parseJSONResponse(content string) ([]ParsedProperty, error) {
	// Strip markdown code blocks if present
	content = strings.TrimSpace(content)
	if strings.HasPrefix(content, "```") {
		lines := strings.Split(content, "\n")
		// Remove first line (```json or ```) and last line (```)
		if len(lines) > 2 {
			lines = lines[1 : len(lines)-1]
		}
		content = strings.Join(lines, "\n")
	}
	content = strings.TrimSpace(content)

	// Try to parse as array
	var properties []ParsedProperty
	if err := json.Unmarshal([]byte(content), &properties); err != nil {
		// Try to parse as single object and wrap in array
		var single ParsedProperty
		if err2 := json.Unmarshal([]byte(content), &single); err2 != nil {
			return nil, fmt.Errorf("failed to parse Gemini JSON response: %w (original: %w)", err2, err)
		}
		properties = []ParsedProperty{single}
	}

	// Filter out empty entries & deduplicate by (property_name + room_number)
	seen := make(map[string]bool)
	var result []ParsedProperty
	for _, p := range properties {
		if p.PropertyName == "" && p.PhoneNumber == "" {
			continue
		}
		// 重複キー: 物件名+号室の組み合わせ（大文字小文字・スペース正規化）
		key := strings.ToLower(strings.TrimSpace(p.PropertyName)) + "|" + strings.TrimSpace(p.RoomNumber)
		if seen[key] {
			log.Printf("⚠️  [dedup] Skipping duplicate: %s 号室%s", p.PropertyName, p.RoomNumber)
			continue
		}
		seen[key] = true
		result = append(result, p)
	}

	log.Printf("✅ Gemini extracted %d properties (after dedup)", len(result))
	return result, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
