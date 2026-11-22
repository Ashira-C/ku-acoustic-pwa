package importer

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Ashira-C/ku-acoustic-pwa/api/internal/db"
)

// Apply writes all members in one transaction: either every row is imported or none is.
// Re-running is safe before launch: users are upserted by student ID and their parts replaced.
func Apply(ctx context.Context, pool *pgxpool.Pool, members []Member) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	// Rollback is a no-op after a successful Commit.
	defer func() { _ = tx.Rollback(ctx) }()

	q := db.New(tx)

	activated, err := q.CountActivatedUsers(ctx)
	if err != nil {
		return fmt.Errorf("count activated users: %w", err)
	}
	if activated > 0 {
		return fmt.Errorf("%d account(s) already activated: the importer is only for the initial load", activated)
	}

	parts, err := q.ListParts(ctx)
	if err != nil {
		return fmt.Errorf("list parts: %w", err)
	}
	partIDs := make(map[string]uuid.UUID, len(parts))
	for _, p := range parts {
		partIDs[p.Name] = p.ID
	}

	for _, m := range members {
		userID, err := q.UpsertImportedUser(ctx, db.UpsertImportedUserParams{
			StudentID: m.StudentID,
			Name:      m.Name,
			Nickname:  m.Nickname,
			Role:      db.UserRole(m.Role),
		})
		if err != nil {
			return fmt.Errorf("upsert %s: %w", m.StudentID, err)
		}

		if err := q.DeleteUserParts(ctx, userID); err != nil {
			return fmt.Errorf("clear parts of %s: %w", m.StudentID, err)
		}
		for _, name := range m.Parts {
			partID, ok := partIDs[name]
			if !ok {
				return fmt.Errorf("%s: part %q is not in the parts table", m.StudentID, name)
			}
			if err := q.AddUserPart(ctx, db.AddUserPartParams{UserID: userID, PartID: partID}); err != nil {
				return fmt.Errorf("add part %s to %s: %w", name, m.StudentID, err)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}