package tenant

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/LewisBz/DirectOrder-Agentic-Platform-DOAP-/internal/platform/database"
)

var ErrNotFound = errors.New("tenant not found")

type Public struct {
	ID           uuid.UUID      `json:"id"`
	Slug         string         `json:"slug"`
	Host         string         `json:"host"`
	Name         string         `json:"name"`
	Currency     string         `json:"currency"`
	TaxName      string         `json:"tax_name"`
	Timezone     string         `json:"timezone"`
	OpeningHours map[string]any `json:"opening_hours,omitempty"`
}

type Resolver interface {
	Resolve(ctx context.Context, host, slug string) (*Public, error)
}

type Store struct {
	db *database.DB
}

func NewStore(db *database.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Resolve(ctx context.Context, host, slug string) (*Public, error) {
	if s == nil || s.db == nil || s.db.Pool == nil {
		return nil, ErrNotFound
	}
	var t Public
	var hours map[string]any
	err := s.db.Pool.QueryRow(ctx, `
		SELECT id, slug, host, name, currency, tax_name, timezone, opening_hours
		FROM resolve_tenant($1, $2)
		LIMIT 1
	`, host, slug).Scan(&t.ID, &t.Slug, &t.Host, &t.Name, &t.Currency, &t.TaxName, &t.Timezone, &hours)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	t.Currency = strings.TrimSpace(t.Currency)
	t.OpeningHours = hours
	return &t, nil
}
