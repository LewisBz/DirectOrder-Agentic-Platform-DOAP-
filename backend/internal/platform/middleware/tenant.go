package middleware

import (
	"context"
	"net"
	"net/http"
	"strings"
)

type ctxKey int

const tenantHostKey ctxKey = 1

func RequestHost(r *http.Request) string {
	h := r.Header.Get("X-Forwarded-Host")
	if h == "" {
		h = r.Host
	}
	if h == "" {
		h = r.Header.Get("Host")
	}
	host, _, err := net.SplitHostPort(h)
	if err != nil {
		host = h
	}
	return strings.ToLower(strings.TrimSpace(host))
}

func IgnoreClientTenantID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Del("X-Tenant-Id")
		host := RequestHost(r)
		ctx := context.WithValue(r.Context(), tenantHostKey, host)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func HostFromContext(ctx context.Context) string {
	h, _ := ctx.Value(tenantHostKey).(string)
	return h
}
