package auth

import (
	"net/http"
	"strings"
)

type GuestSessionPublic struct {
	Token      string `json:"token"`
	TenantSlug string `json:"tenant_slug"`
	Channel    string `json:"channel"`
}

func guestTokenFromRequest(r *http.Request, bodyToken string) string {
	raw := strings.TrimSpace(r.Header.Get("X-Guest-Token"))
	if raw == "" {
		raw = strings.TrimSpace(bodyToken)
	}
	if raw == "" {
		if c, err := r.Cookie("doap_guest"); err == nil {
			raw = strings.TrimSpace(c.Value)
		}
	}
	return raw
}
