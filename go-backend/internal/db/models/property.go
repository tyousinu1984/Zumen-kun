package models

import "time"

type Property struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	RoomNumber   string    `json:"room_number"`
	PhoneNumber  string    `json:"phone_number"`
	Price        float64   `json:"price"`
	AdRate       string    `json:"ad_rate"`
	Status       string    `json:"status"` // pending, parsing, calling, completed
	ForeignerOK  *bool     `json:"foreigner_ok"`
	PDFURL       string    `json:"pdf_url"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
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
