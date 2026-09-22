package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type staffCtxKey int

const staffClaimsKey staffCtxKey = 1

type StaffClaims struct {
	StaffID  uuid.UUID `json:"sub_uuid"`
	TenantID uuid.UUID `json:"tenant_id"`
	Role     string    `json:"role"`
	jwt.RegisteredClaims
}

func (c StaffClaims) ValidRole() bool {
	switch c.Role {
	case "owner", "cashier", "kitchen":
		return true
	default:
		return false
	}
}

func StaffFromContext(ctx context.Context) (StaffClaims, bool) {
	c, ok := ctx.Value(staffClaimsKey).(StaffClaims)
	return c, ok
}

// RequireStaff parses Bearer HS256 access JWT. lookupTenant returns the commerce id for RequestHost.
// jwt.tenant_id must equal that id.
func RequireStaff(secret string, lookupTenant func(*http.Request) (uuid.UUID, error)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw := strings.TrimSpace(r.Header.Get("Authorization"))
			const prefix = "Bearer "
			if !strings.HasPrefix(raw, prefix) {
				writeAuthJSON(w, http.StatusUnauthorized, "invalid_credentials", "authentication required")
				return
			}
			tokenStr := strings.TrimSpace(strings.TrimPrefix(raw, prefix))
			var claims StaffClaims
			tok, err := jwt.ParseWithClaims(tokenStr, &claims, func(t *jwt.Token) (any, error) {
				if t.Method != jwt.SigningMethodHS256 {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(secret), nil
			})
			if err != nil || !tok.Valid {
				writeAuthJSON(w, http.StatusUnauthorized, "invalid_credentials", "authentication required")
				return
			}
			if claims.StaffID == uuid.Nil {
				if sub, e := uuid.Parse(claims.Subject); e == nil {
					claims.StaffID = sub
				}
			}
			if claims.StaffID == uuid.Nil || claims.TenantID == uuid.Nil || !claims.ValidRole() {
				writeAuthJSON(w, http.StatusUnauthorized, "invalid_credentials", "authentication required")
				return
			}
			hostTenant, err := lookupTenant(r)
			if err != nil || hostTenant != claims.TenantID {
				writeAuthJSON(w, http.StatusUnauthorized, "invalid_credentials", "authentication required")
				return
			}
			ctx := context.WithValue(r.Context(), staffClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func writeAuthJSON(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": code, "message": message})
}
