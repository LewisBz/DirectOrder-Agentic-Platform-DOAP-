package auth

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/LewisBz/DirectOrder-Agentic-Platform-DOAP-/internal/platform/database"
)

var (
	isoTenantA = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	isoTenantB = uuid.MustParse("22222222-2222-2222-2222-222222222222")
)

func openAppDB(t *testing.T) *database.DB {
	t.Helper()
	host := envDefault("POSTGRES_HOST", "localhost")
	port := envDefault("POSTGRES_PORT", "5432")
	dbn := envDefault("POSTGRES_DB", "doap")
	user := envDefault("POSTGRES_APP_USER", "doap_app")
	pass := envDefault("POSTGRES_APP_PASSWORD", "changeme_app")
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, dbn)
	db, err := database.Open(context.Background(), dsn)
	if err != nil {
		t.Skipf("isolation tests need Postgres as doap_app (docker compose up): %v", err)
	}
	t.Cleanup(db.Close)
	return db
}

func envDefault(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func TestStaffACannotSelectOrUpdateStaffOfB(t *testing.T) {
	db := openAppDB(t)
	ctx := context.Background()
	err := db.WithTenantTx(ctx, isoTenantA, func(tx pgx.Tx) error {
		var n int
		if err := tx.QueryRow(ctx, `SELECT count(*) FROM staff WHERE tenant_id = $1`, isoTenantB).Scan(&n); err != nil {
			return err
		}
		if n != 0 {
			t.Errorf("tenant A saw %d staff rows of B", n)
		}
		tag, err := tx.Exec(ctx, `UPDATE staff SET name = 'hacked' WHERE tenant_id = $1`, isoTenantB)
		if err != nil {
			return err
		}
		if tag.RowsAffected() != 0 {
			t.Errorf("tenant A updated %d staff rows of B", tag.RowsAffected())
		}
		if err := tx.QueryRow(ctx, `SELECT count(*) FROM staff_refresh_tokens WHERE tenant_id = $1`, isoTenantB).Scan(&n); err != nil {
			return err
		}
		if n != 0 {
			t.Errorf("tenant A saw %d refresh rows of B", n)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestStaffNoSettingYieldsZeroRows(t *testing.T) {
	db := openAppDB(t)
	ctx := context.Background()
	var n int
	if err := db.Pool.QueryRow(ctx, `SELECT count(*) FROM staff`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("without app.tenant_id expected 0 staff rows, got %d", n)
	}
	if err := db.Pool.QueryRow(ctx, `SELECT count(*) FROM staff_refresh_tokens`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("without app.tenant_id expected 0 refresh rows, got %d", n)
	}
}

func TestStoreQueriesStayInsideTenant(t *testing.T) {
	db := openAppDB(t)
	ctx := context.Background()
	store := NewStore(db)
	rows, err := store.List(ctx, isoTenantA)
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		if row.TenantID != isoTenantA {
			t.Fatalf("list leaked tenant %s", row.TenantID)
		}
		if row.Email == "owner@demo-b.local" {
			t.Fatal("list included Demo B owner")
		}
	}

	var bID uuid.UUID
	err = db.WithTenantTx(ctx, isoTenantB, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `SELECT id FROM staff WHERE email = 'owner@demo-b.local'`).Scan(&bID)
	})
	if err != nil {
		t.Skipf("seed owner@demo-b.local required: %v", err)
	}
	_, err = store.GetByID(ctx, isoTenantA, bID)
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("GetByID B as A: %v", err)
	}
	err = store.SetActive(ctx, isoTenantA, bID, false)
	if !errors.Is(err, ErrStaffNotFound) {
		t.Fatalf("SetActive B as A: %v", err)
	}
}
