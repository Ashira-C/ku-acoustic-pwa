package part

import (
	"context"
	"fmt"

	"github.com/Ashira-C/ku-acoustic-pwa/api/internal/db"
)

// Repository is the only layer that talks to sqlc.
type Repository struct {
	q *db.Queries
}

func NewRepository(q *db.Queries) *Repository {
	return &Repository{q: q}
}

func (r *Repository) List(ctx context.Context) ([]Part, error) {
	rows, err := r.q.ListParts(ctx)
	if err != nil {
		return nil, fmt.Errorf("list parts: %w", err)
	}

	parts := make([]Part, 0, len(rows))
	for _, row := range rows {
		parts = append(parts, Part{
			ID:           row.ID,
			Name:         row.Name,
			DisplayOrder: row.DisplayOrder,
		})
	}
	return parts, nil
}