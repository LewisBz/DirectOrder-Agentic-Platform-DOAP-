package auth

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	platmw "github.com/LewisBz/DirectOrder-Agentic-Platform-DOAP-/internal/platform/middleware"
	"github.com/LewisBz/DirectOrder-Agentic-Platform-DOAP-/internal/tenant"
)

type Service struct {
	dir     Directory
	tenants tenant.Resolver
	secret  string
	limit   *LoginLimiter
}

func NewService(dir Directory, tenants tenant.Resolver, secret string) *Service {
	return &Service{dir: dir, tenants: tenants, secret: secret, limit: NewLoginLimiter()}
}

func (s *Service) resolveTenant(r *http.Request) (*tenant.Public, error) {
	if s == nil || s.tenants == nil {
		return nil, tenant.ErrNotFound
	}
	return s.tenants.Resolve(r.Context(), platmw.HostFromContext(r.Context()), "")
}

func (s *Service) pair(ctx context.Context, staff *StaffRow) (*TokenPair, error) {
	access, err := IssueAccess(s.secret, staff.ID, staff.TenantID, staff.Role)
	if err != nil {
		return nil, err
	}
	raw, hash, err := NewRefreshToken()
	if err != nil {
		return nil, err
	}
	if err := s.dir.InsertRefresh(ctx, staff.TenantID, staff.ID, hash, time.Now().Add(RefreshTTL)); err != nil {
		return nil, err
	}
	return &TokenPair{
		AccessToken:  access,
		RefreshToken: raw,
		ExpiresIn:    int(AccessTTL.Seconds()),
		TokenType:    "Bearer",
	}, nil
}

func Login(s *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeErr(w, http.StatusBadRequest, "validation_error", "invalid json")
			return
		}
		body.Email = strings.ToLower(strings.TrimSpace(body.Email))
		if body.Email == "" || body.Password == "" {
			writeErr(w, http.StatusBadRequest, "validation_error", "email and password required")
			return
		}
		tn, err := s.resolveTenant(r)
		if err != nil {
			writeErr(w, http.StatusUnauthorized, "invalid_credentials", "invalid email or password")
			return
		}
		ip := r.RemoteAddr
		if !s.limit.Allow(ip, tn.ID.String(), body.Email) {
			writeErr(w, http.StatusTooManyRequests, "rate_limited", "invalid email or password")
			return
		}
		staff, err := s.dir.FindByEmail(r.Context(), tn.ID, body.Email)
		if err != nil || staff == nil || !staff.Active {
			writeErr(w, http.StatusUnauthorized, "invalid_credentials", "invalid email or password")
			return
		}
		if err := ComparePassword(staff.PasswordHash, body.Password); err != nil {
			writeErr(w, http.StatusUnauthorized, "invalid_credentials", "invalid email or password")
			return
		}
		pair, err := s.pair(r.Context(), staff)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, "validation_error", err.Error())
			return
		}
		writeJSON(w, http.StatusOK, pair)
	}
}

func Refresh(s *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			RefreshToken string `json:"refresh_token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.RefreshToken) == "" {
			writeErr(w, http.StatusUnauthorized, "invalid_credentials", "invalid email or password")
			return
		}
		tn, err := s.resolveTenant(r)
		if err != nil {
			writeErr(w, http.StatusUnauthorized, "invalid_credentials", "invalid email or password")
			return
		}
		hash := HashRefresh(body.RefreshToken)
		staffID, exp, revoked, err := s.dir.LookupRefresh(r.Context(), tn.ID, hash)
		if err != nil || revoked || time.Now().After(exp) {
			writeErr(w, http.StatusUnauthorized, "invalid_credentials", "invalid email or password")
			return
		}
		staff, err := s.dir.GetByID(r.Context(), tn.ID, staffID)
		if err != nil || staff == nil || !staff.Active {
			writeErr(w, http.StatusUnauthorized, "invalid_credentials", "invalid email or password")
			return
		}
		_ = s.dir.RevokeRefresh(r.Context(), tn.ID, hash)
		pair, err := s.pair(r.Context(), staff)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, "validation_error", err.Error())
			return
		}
		writeJSON(w, http.StatusOK, pair)
	}
}

func Logout(s *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := platmw.StaffFromContext(r.Context())
		if !ok {
			writeErr(w, http.StatusUnauthorized, "invalid_credentials", "invalid email or password")
			return
		}
		_ = s.dir.RevokeAllRefresh(r.Context(), claims.TenantID, claims.StaffID)
		w.WriteHeader(http.StatusNoContent)
	}
}

func Me(s *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := platmw.StaffFromContext(r.Context())
		if !ok {
			writeErr(w, http.StatusUnauthorized, "invalid_credentials", "invalid email or password")
			return
		}
		tn, err := s.resolveTenant(r)
		if err != nil || tn.ID != claims.TenantID {
			writeErr(w, http.StatusUnauthorized, "invalid_credentials", "invalid email or password")
			return
		}
		staff, err := s.dir.GetByID(r.Context(), tn.ID, claims.StaffID)
		if err != nil || staff == nil || !staff.Active {
			writeErr(w, http.StatusUnauthorized, "invalid_credentials", "invalid email or password")
			return
		}
		writeJSON(w, http.StatusOK, StaffPublic{
			ID:         staff.ID,
			Email:      staff.Email,
			Name:       staff.Name,
			Role:       staff.Role,
			Active:     staff.Active,
			TenantSlug: tn.Slug,
			TenantName: tn.Name,
		})
	}
}

func EnsureGuest(s *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tn, err := s.resolveTenant(r)
		if err != nil {
			writeErr(w, http.StatusNotFound, "tenant_not_found", "unknown host or slug")
			return
		}
		var body struct {
			Channel string `json:"channel"`
			Token   string `json:"token"`
		}
		if r.Body != nil {
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil && !errors.Is(err, io.EOF) {
				writeErr(w, http.StatusBadRequest, "validation_error", "invalid json")
				return
			}
		}
		channel := strings.TrimSpace(body.Channel)
		if channel == "" {
			channel = "pwa"
		}
		if channel != "pwa" {
			writeErr(w, http.StatusUnprocessableEntity, "validation_error", "channel must be pwa")
			return
		}
		raw := guestTokenFromRequest(r, body.Token)
		if raw != "" && s.dir != nil {
			hash := HashRefresh(raw)
			ok, err := s.dir.FindGuest(r.Context(), tn.ID, hash)
			if err != nil {
				writeErr(w, http.StatusInternalServerError, "validation_error", err.Error())
				return
			}
			if ok {
				_ = s.dir.TouchGuest(r.Context(), tn.ID, hash)
				writeJSON(w, http.StatusOK, GuestSessionPublic{Token: raw, TenantSlug: tn.Slug, Channel: channel})
				return
			}
		}
		newRaw, hash, err := NewRefreshToken()
		if err != nil {
			writeErr(w, http.StatusInternalServerError, "validation_error", err.Error())
			return
		}
		if s.dir == nil {
			writeErr(w, http.StatusInternalServerError, "validation_error", "guest store unavailable")
			return
		}
		if err := s.dir.InsertGuest(r.Context(), tn.ID, hash, channel); err != nil {
			writeErr(w, http.StatusInternalServerError, "validation_error", err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, GuestSessionPublic{Token: newRaw, TenantSlug: tn.Slug, Channel: channel})
	}
}

func LookupTenantID(tenants tenant.Resolver) func(*http.Request) (uuid.UUID, error) {
	return func(r *http.Request) (uuid.UUID, error) {
		if tenants == nil {
			return uuid.Nil, errors.New("no tenant resolver")
		}
		r.Header.Del("X-Tenant-Id")
		tn, err := tenants.Resolve(r.Context(), platmw.HostFromContext(r.Context()), "")
		if err != nil {
			return uuid.Nil, err
		}
		return tn.ID, nil
	}
}

func requireOwner(w http.ResponseWriter, r *http.Request) (platmw.StaffClaims, bool) {
	claims, ok := platmw.StaffFromContext(r.Context())
	if !ok {
		writeErr(w, http.StatusUnauthorized, "invalid_credentials", "invalid email or password")
		return claims, false
	}
	if claims.Role != "owner" {
		writeErr(w, http.StatusForbidden, "forbidden", "not allowed")
		return claims, false
	}
	return claims, true
}

func (s *Service) publicFrom(tn *tenant.Public, row *StaffRow) StaffPublic {
	return StaffPublic{
		ID:         row.ID,
		Email:      row.Email,
		Name:       row.Name,
		Role:       row.Role,
		Active:     row.Active,
		TenantSlug: tn.Slug,
		TenantName: tn.Name,
	}
}

func ListStaff(s *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireOwner(w, r); !ok {
			return
		}
		tn, err := s.resolveTenant(r)
		if err != nil {
			writeErr(w, http.StatusUnauthorized, "invalid_credentials", "invalid email or password")
			return
		}
		rows, err := s.dir.List(r.Context(), tn.ID)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, "validation_error", err.Error())
			return
		}
		out := make([]StaffPublic, 0, len(rows))
		for i := range rows {
			out = append(out, s.publicFrom(tn, &rows[i]))
		}
		writeJSON(w, http.StatusOK, out)
	}
}

func CreateStaff(s *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireOwner(w, r); !ok {
			return
		}
		var body struct {
			Email    string `json:"email"`
			Password string `json:"password"`
			Name     string `json:"name"`
			Role     string `json:"role"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeErr(w, http.StatusBadRequest, "validation_error", "invalid json")
			return
		}
		body.Email = strings.ToLower(strings.TrimSpace(body.Email))
		body.Name = strings.TrimSpace(body.Name)
		if body.Email == "" || body.Name == "" || len(body.Password) < 8 {
			writeErr(w, http.StatusUnprocessableEntity, "validation_error", "email, name and password (min 8) required")
			return
		}
		hash, err := HashPassword(body.Password)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, "validation_error", err.Error())
			return
		}
		tn, err := s.resolveTenant(r)
		if err != nil {
			writeErr(w, http.StatusUnauthorized, "invalid_credentials", "invalid email or password")
			return
		}
		row, err := s.dir.Create(r.Context(), tn.ID, body.Email, body.Name, body.Role, hash)
		if err != nil {
			writeStaffErr(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(s.publicFrom(tn, row))
	}
}

func PatchStaff(s *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireOwner(w, r); !ok {
			return
		}
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			writeStaffErr(w, ErrStaffNotFound)
			return
		}
		var body struct {
			Email    *string `json:"email"`
			Password *string `json:"password"`
			Name     *string `json:"name"`
			Role     *string `json:"role"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeErr(w, http.StatusBadRequest, "validation_error", "invalid json")
			return
		}
		tn, err := s.resolveTenant(r)
		if err != nil {
			writeErr(w, http.StatusUnauthorized, "invalid_credentials", "invalid email or password")
			return
		}
		cur, err := s.dir.GetByID(r.Context(), tn.ID, id)
		if err != nil {
			writeStaffErr(w, ErrStaffNotFound)
			return
		}
		email, name, role := cur.Email, cur.Name, cur.Role
		if body.Email != nil {
			email = strings.ToLower(strings.TrimSpace(*body.Email))
		}
		if body.Name != nil {
			name = strings.TrimSpace(*body.Name)
		}
		if body.Role != nil {
			role = *body.Role
		}
		var hash *string
		if body.Password != nil {
			if len(*body.Password) < 8 {
				writeErr(w, http.StatusUnprocessableEntity, "validation_error", "password min 8")
				return
			}
			h, err := HashPassword(*body.Password)
			if err != nil {
				writeErr(w, http.StatusInternalServerError, "validation_error", err.Error())
				return
			}
			hash = &h
		}
		row, err := s.dir.Update(r.Context(), tn.ID, id, email, name, role, hash)
		if err != nil {
			writeStaffErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, s.publicFrom(tn, row))
	}
}

func DeactivateStaff(s *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		setStaffActive(s, w, r, false)
	}
}

func ReactivateStaff(s *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		setStaffActive(s, w, r, true)
	}
}

func setStaffActive(s *Service, w http.ResponseWriter, r *http.Request, active bool) {
	if _, ok := requireOwner(w, r); !ok {
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeStaffErr(w, ErrStaffNotFound)
		return
	}
	tn, err := s.resolveTenant(r)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, "invalid_credentials", "invalid email or password")
		return
	}
	if err := s.dir.SetActive(r.Context(), tn.ID, id, active); err != nil {
		writeStaffErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeStaffErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrConflict):
		writeErr(w, http.StatusConflict, "conflict", "email already used in this commerce")
	case errors.Is(err, ErrOwnerRole), errors.Is(err, ErrOwnerLocked):
		writeErr(w, http.StatusUnprocessableEntity, "validation_error", err.Error())
	case errors.Is(err, ErrStaffNotFound):
		writeErr(w, http.StatusNotFound, "not_found", "staff not found")
	default:
		writeErr(w, http.StatusInternalServerError, "validation_error", err.Error())
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]string{"error": code, "message": message})
}
