package tenant

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/LewisBz/DirectOrder-Agentic-Platform-DOAP-/internal/platform/database"
)

var (
	tenantA = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	tenantB = uuid.MustParse("22222222-2222-2222-2222-222222222222")
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

func TestAppRoleIsNotTableOwner(t *testing.T) {
	db := openAppDB(t)
	var owner, me string
	if err := db.Pool.QueryRow(context.Background(), `SELECT tableowner FROM pg_tables WHERE schemaname = 'public' AND tablename = 'tenants'`).Scan(&owner); err != nil {
		t.Fatal(err)
	}
	if err := db.Pool.QueryRow(context.Background(), `SELECT current_user`).Scan(&me); err != nil {
		t.Fatal(err)
	}
	if me == owner {
		t.Fatalf("app role %s must not own table tenants (owner=%s)", me, owner)
	}
	if owner != "doap_owner" {
		t.Fatalf("expected table owner doap_owner, got %s", owner)
	}
}

func TestIsolationCanaryACannotReadOrUpdateB(t *testing.T) {
	db := openAppDB(t)
	ctx := context.Background()
	err := db.WithTenantTx(ctx, tenantA, func(tx pgx.Tx) error {
		var n int
		if err := tx.QueryRow(ctx, `SELECT count(*) FROM isolation_canaries WHERE tenant_id = $1`, tenantB).Scan(&n); err != nil {
			return err
		}
		if n != 0 {
			t.Errorf("tenant A saw %d canary rows of B", n)
		}
		tag, err := tx.Exec(ctx, `UPDATE isolation_canaries SET label = 'hacked' WHERE tenant_id = $1`, tenantB)
		if err != nil {
			return err
		}
		if tag.RowsAffected() != 0 {
			t.Errorf("tenant A updated %d canary rows of B", tag.RowsAffected())
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestIsolationNoSettingYieldsZeroBusinessRows(t *testing.T) {
	db := openAppDB(t)
	var n int
	if err := db.Pool.QueryRow(context.Background(), `SELECT count(*) FROM isolation_canaries`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("without app.tenant_id expected 0 canary rows, got %d", n)
	}
}
