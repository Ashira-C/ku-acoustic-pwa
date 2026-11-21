package part

import "github.com/google/uuid"

// Part is the domain type returned by the API (decoupled from sqlc's db.Part).
type Part struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	DisplayOrder int16     `json:"displayOrder"`
}