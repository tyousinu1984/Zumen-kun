package models

import "time"

type Property struct {
	ID          string     `json:"id,omitempty"`
	CompanyID   string     `json:"company_id,omitempty"`
	Name        string     `json:"name,omitempty"`
	RoomNumber  string     `json:"room_number,omitempty"`
	PhoneNumber string     `json:"phone_number,omitempty"`
	Price       float64    `json:"price,omitempty"`
	AdRate      string     `json:"ad_rate,omitempty"`
	Status      string     `json:"status,omitempty"`
	ForeignerOK *bool      `json:"foreigner_ok,omitempty"`
	PDFURL      string     `json:"pdf_url,omitempty"`
	SourcePDFID string     `json:"source_pdf_id,omitempty"` // 元PDFのpropertyID
	PageNumber  int        `json:"page_number,omitempty"`   // 何ページ目か
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

type CreatePropertyRequest struct {
	Name        string  `json:"name"`
	RoomNumber  string  `json:"room_number"`
	PhoneNumber string  `json:"phone_number"`
	Price       float64 `json:"price"`
	AdRate      string  `json:"ad_rate"`
	PDFURL      string  `json:"pdf_url"`
}

type UpdatePropertyRequest struct {
	Name        *string  `json:"name,omitempty"`
	RoomNumber  *string  `json:"room_number,omitempty"`
	PhoneNumber *string  `json:"phone_number,omitempty"`
	Price       *float64 `json:"price,omitempty"`
	AdRate      *string  `json:"ad_rate,omitempty"`
	Status      *string  `json:"status,omitempty"`
	ForeignerOK *bool    `json:"foreigner_ok,omitempty"`
}
