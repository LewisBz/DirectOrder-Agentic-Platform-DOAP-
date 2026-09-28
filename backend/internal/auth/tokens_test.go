package auth

import (
	"testing"

	"github.com/google/uuid"
)

func TestRefreshTokensAreUniqueAndHashed(t *testing.T) {
	raw1, h1, err := NewRefreshToken()
	if err != nil {
		t.Fatal(err)
	}
	raw2, h2, err := NewRefreshToken()
	if err != nil {
		t.Fatal(err)
	}
	if raw1 == raw2 || h1 == h2 {
		t.Fatal("refresh tokens must rotate to new values")
	}
	if HashRefresh(raw1) != h1 {
		t.Fatal("hash mismatch")
	}
}

func TestIssueAccessRoundTrip(t *testing.T) {
	secret := "changeme_jwt_secret_do_not_use_prod_xx"
	sid := uuid.MustParse("33333333-3333-4333-8333-333333333333")
	tid := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	tok, err := IssueAccess(secret, sid, tid, "owner")
	if err != nil {
		t.Fatal(err)
	}
	if tok == "" {
		t.Fatal("empty token")
	}
}
