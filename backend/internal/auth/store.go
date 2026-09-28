package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/LewisBz/DirectOrder-Agentic-Platform-DOAP-/internal/platform/database"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInactive           = errors.New("inactive")
	ErrConflict           = errors.New("conflict")
	ErrStaffNotFound      = errors.New("staff not found")
	ErrOwnerRole          = errors.New("owner role not allowed")
	ErrOwnerLocked        = errors.New("owner cannot be mutated")
)

type StaffRow struct {
	ID           uuid.UUID
	TenantID     uuid.UUID
	Email        string
	Name         string
	Role         string
	PasswordHash string
	Active       bool
}

type StaffPublic struct {
	ID         uuid.UUID `json:"id"`
	Email      string    `json:"email"`
	Name       string    `json:"name"`
	Role       string    `json:"role"`
	Active     bool      `json:"active"`
	TenantSlug string    `json:"tenant_slug"`
	TenantName string    `json:"tenant_name"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type Directory interface {
	FindByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*StaffRow, error)
	GetByID(ctx context.Context, tenantID, staffID uuid.UUID) (*StaffRow, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]StaffRow, error)
	Create(ctx context.Context, tenantID uuid.UUID, email, name, role, passwordHash string) (*StaffRow, error)
	Update(ctx context.Context, tenantID, staffID uuid.UUID, email, name, role string, passwordHash *string) (*StaffRow, error)
	SetActive(ctx context.Context, tenantID, staffID uuid.UUID, active bool) error
	InsertRefresh(ctx context.Context, tenantID, staffID uuid.UUID, tokenHash string, expiresAt time.Time) error
	LookupRefresh(ctx context.Context, tenantID uuid.UUID, tokenHash string) (staffID uuid.UUID, expiresAt time.Time, revoked bool, err error)
	RevokeRefresh(ctx context.Context, tenantID uuid.UUID, tokenHash string) error
	RevokeAllRefresh(ctx context.Context, tenantID, staffID uuid.UUID) error
	FindGuest(ctx context.Context, tenantID uuid.UUID, tokenHash string) (bool, error)
	InsertGuest(ctx context.Context, tenantID uuid.UUID, tokenHash, channel string) error
	TouchGuest(ctx context.Context, tenantID uuid.UUID, tokenHash string) error
}

// Store runs every staff, refresh, and guest query inside WithTenantTx (SET LOCAL app.tenant_id).
type Store struct {
	db *database.DB
}

func NewStore(db *database.DB) *Store {
	return &Store{db: db}
}

func (s *Store) FindByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*StaffRow, error) {
	if s == nil || s.db == nil || s.db.Pool == nil {
		return nil, ErrInvalidCredentials
	}
	email = strings.ToLower(strings.TrimSpace(email))
	var row StaffRow
	err := s.db.WithTenantTx(ctx, tenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `
			SELECT id, tenant_id, email, name, role, password_hash, active
			FROM staff WHERE email = $1
		`, email).Scan(&row.ID, &row.TenantID, &row.Email, &row.Name, &row.Role, &row.PasswordHash, &row.Active)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (s *Store) GetByID(ctx context.Context, tenantID, staffID uuid.UUID) (*StaffRow, error) {
	var row StaffRow
	err := s.db.WithTenantTx(ctx, tenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `
			SELECT id, tenant_id, email, name, role, password_hash, active
			FROM staff WHERE id = $1
		`, staffID).Scan(&row.ID, &row.TenantID, &row.Email, &row.Name, &row.Role, &row.PasswordHash, &row.Active)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrInvalidCredentials
	}
	return &row, err
}

func (s *Store) InsertRefresh(ctx context.Context, tenantID, staffID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	return s.db.WithTenantTx(ctx, tenantID, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			INSERT INTO staff_refresh_tokens (tenant_id, staff_id, token_hash, expires_at)
			VALUES ($1, $2, $3, $4)
		`, tenantID, staffID, tokenHash, expiresAt)
		return err
	})
}

func (s *Store) LookupRefresh(ctx context.Context, tenantID uuid.UUID, tokenHash string) (uuid.UUID, time.Time, bool, error) {
	var staffID uuid.UUID
	var exp time.Time
	var revokedAt *time.Time
	err := s.db.WithTenantTx(ctx, tenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `
			SELECT staff_id, expires_at, revoked_at
			FROM staff_refresh_tokens WHERE token_hash = $1
		`, tokenHash).Scan(&staffID, &exp, &revokedAt)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, time.Time{}, false, ErrInvalidCredentials
	}
	if err != nil {
		return uuid.Nil, time.Time{}, false, err
	}
	return staffID, exp, revokedAt != nil, nil
}

func (s *Store) RevokeRefresh(ctx context.Context, tenantID uuid.UUID, tokenHash string) error {
	return s.db.WithTenantTx(ctx, tenantID, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			UPDATE staff_refresh_tokens SET revoked_at = now()
			WHERE token_hash = $1 AND revoked_at IS NULL
		`, tokenHash)
		return err
	})
}

func (s *Store) RevokeAllRefresh(ctx context.Context, tenantID, staffID uuid.UUID) error {
	return s.db.WithTenantTx(ctx, tenantID, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			UPDATE staff_refresh_tokens SET revoked_at = now()
			WHERE staff_id = $1 AND revoked_at IS NULL
		`, staffID)
		return err
	})
}

func operableRole(role string) bool {
	return role == "cashier" || role == "kitchen"
}

func (s *Store) List(ctx context.Context, tenantID uuid.UUID) ([]StaffRow, error) {
	if s == nil || s.db == nil || s.db.Pool == nil {
		return nil, ErrInvalidCredentials
	}
	var out []StaffRow
	err := s.db.WithTenantTx(ctx, tenantID, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, `
			SELECT id, tenant_id, email, name, role, password_hash, active
			FROM staff ORDER BY created_at
		`)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var row StaffRow
			if err := rows.Scan(&row.ID, &row.TenantID, &row.Email, &row.Name, &row.Role, &row.PasswordHash, &row.Active); err != nil {
				return err
			}
			out = append(out, row)
		}
		return rows.Err()
	})
	return out, err
}

func (s *Store) Create(ctx context.Context, tenantID uuid.UUID, email, name, role, passwordHash string) (*StaffRow, error) {
	if !operableRole(role) {
		return nil, ErrOwnerRole
	}
	email = strings.ToLower(strings.TrimSpace(email))
	var row StaffRow
	err := s.db.WithTenantTx(ctx, tenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `
			INSERT INTO staff (tenant_id, email, name, role, password_hash, active)
			VALUES ($1, $2, $3, $4, $5, true)
			RETURNING id, tenant_id, email, name, role, password_hash, active
		`, tenantID, email, name, role, passwordHash).Scan(
			&row.ID, &row.TenantID, &row.Email, &row.Name, &row.Role, &row.PasswordHash, &row.Active,
		)
	})
	if isUnique(err) {
		return nil, ErrConflict
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (s *Store) Update(ctx context.Context, tenantID, staffID uuid.UUID, email, name, role string, passwordHash *string) (*StaffRow, error) {
	cur, err := s.GetByID(ctx, tenantID, staffID)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			return nil, ErrStaffNotFound
		}
		return nil, err
	}
	if cur.Role == "owner" {
		return nil, ErrOwnerLocked
	}
	if !operableRole(role) {
		return nil, ErrOwnerRole
	}
	email = strings.ToLower(strings.TrimSpace(email))
	hash := cur.PasswordHash
	if passwordHash != nil {
		hash = *passwordHash
	}
	var row StaffRow
	err = s.db.WithTenantTx(ctx, tenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `
			UPDATE staff SET email = $2, name = $3, role = $4, password_hash = $5, updated_at = now()
			WHERE id = $1
			RETURNING id, tenant_id, email, name, role, password_hash, active
		`, staffID, email, name, role, hash).Scan(
			&row.ID, &row.TenantID, &row.Email, &row.Name, &row.Role, &row.PasswordHash, &row.Active,
		)
	})
	if isUnique(err) {
		return nil, ErrConflict
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrStaffNotFound
	}
	if err != nil {
		return nil, err
	}
	if passwordHash != nil {
		_ = s.RevokeAllRefresh(ctx, tenantID, staffID)
	}
	return &row, nil
}

func (s *Store) SetActive(ctx context.Context, tenantID, staffID uuid.UUID, active bool) error {
	cur, err := s.GetByID(ctx, tenantID, staffID)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			return ErrStaffNotFound
		}
		return err
	}
	if cur.Role == "owner" {
		return ErrOwnerLocked
	}
	err = s.db.WithTenantTx(ctx, tenantID, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `UPDATE staff SET active = $2, updated_at = now() WHERE id = $1`, staffID, active)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return ErrStaffNotFound
		}
		return nil
	})
	if err != nil {
		return err
	}
	if !active {
		_ = s.RevokeAllRefresh(ctx, tenantID, staffID)
	}
	return nil
}

func isUnique(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "staff_tenant_id_email") || strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "23505")
}

func (s *Store) FindGuest(ctx context.Context, tenantID uuid.UUID, tokenHash string) (bool, error) {
	if s == nil || s.db == nil || s.db.Pool == nil {
		return false, nil
	}
	var n int
	err := s.db.WithTenantTx(ctx, tenantID, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `SELECT count(*) FROM guest_sessions WHERE token_hash = $1`, tokenHash).Scan(&n)
	})
	return n > 0, err
}

func (s *Store) InsertGuest(ctx context.Context, tenantID uuid.UUID, tokenHash, channel string) error {
	if channel == "" {
		channel = "pwa"
	}
	return s.db.WithTenantTx(ctx, tenantID, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			INSERT INTO guest_sessions (tenant_id, token_hash, channel)
			VALUES ($1, $2, $3)
		`, tenantID, tokenHash, channel)
		return err
	})
}

func (s *Store) TouchGuest(ctx context.Context, tenantID uuid.UUID, tokenHash string) error {
	return s.db.WithTenantTx(ctx, tenantID, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `UPDATE guest_sessions SET last_seen = now() WHERE token_hash = $1`, tokenHash)
		return err
	})
}
