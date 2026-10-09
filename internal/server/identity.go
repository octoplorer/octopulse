package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/octoplorer/octopulse/internal/domain"
	"github.com/octoplorer/octopulse/internal/security"
	"github.com/octoplorer/octopulse/internal/store"
)

type SessionBody struct {
	User      *domain.User `json:"user"`
	CSRFToken string       `json:"csrfToken"`
}

func (SessionBody) Schema(registry huma.Registry) *huma.Schema {
	return &huma.Schema{
		Type: "object",
		Properties: map[string]*huma.Schema{
			"user": {
				AnyOf: []*huma.Schema{
					registry.Schema(reflect.TypeFor[domain.User](), true, "User"),
					{
						Type: "null",
					},
				},
			},
			"csrfToken": {
				Type: "string",
			},
		},
		Required: []string{"user", "csrfToken"},
	}
}

type SessionOutput struct {
	SetCookie string `header:"Set-Cookie"`
	Body      SessionBody
}
type Credentials struct {
	Username string `json:"username" minLength:"1" maxLength:"100"`
	Password string `json:"password" minLength:"1" maxLength:"72"`
}
type SetupBody struct {
	Credentials
	OrganizationName string `json:"organizationName"`
	Timezone         string `json:"timezone"`
}
type SetupInput struct{ Body SetupBody }
type SetupStatus struct {
	Required bool `json:"required"`
}
type UserWrite struct {
	Username string      `json:"username"`
	Name     string      `json:"name"`
	Role     domain.Role `json:"role" enum:"admin,operator,viewer"`
	Locale   string      `json:"locale" enum:"zh-CN,en"`
	Timezone string      `json:"timezone"`
	Enabled  *bool       `json:"enabled"`
	Password string      `json:"password,omitempty" maxLength:"72"`
}

func (s *Server) createSession(ctx context.Context, t *store.Tx, u domain.User) (*SessionOutput, error) {
	token := security.Token()
	sess := domain.Session{
		ID:        security.HashToken(token),
		UserID:    u.ID,
		CSRFToken: security.Token(),
		CreatedAt: domain.Now(),
		ExpiresAt: domain.Now() + sessionLifetime.Milliseconds(),
	}
	if e := t.Put(
		ctx,
		"sessions",
		sess.ID,
		sess,
	); e != nil {
		return nil, e
	}
	cookie := http.Cookie{
		Name:     "octopulse_session",
		Value:    token,
		Path:     "/api/v1",
		HttpOnly: true,
		Secure:   s.Config.CookieSecure,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(sessionLifetime.Seconds()),
	}
	return &SessionOutput{SetCookie: cookie.String(), Body: SessionBody{User: &u, CSRFToken: sess.CSRFToken}}, nil
}

func (s *Server) registerIdentity() {
	s.registerProfile()
	huma.Register(
		s.API,
		huma.Operation{
			OperationID: "getSetup",
			Method:      "GET",
			Path:        "/api/v1/setup",
		},
		func(ctx context.Context, _ *struct{}) (*Output[SetupStatus], error) {
			users, e := s.Store.List(ctx, "users")
			if e != nil {
				return nil, apiError(ctx, e)
			}
			return &Output[SetupStatus]{Body: SetupStatus{Required: len(users) == 0}}, nil
		},
	)
	huma.Register(
		s.API,
		huma.Operation{
			OperationID: "createSetup",
			Method:      "POST",
			Path:        "/api/v1/setup",
		},
		func(ctx context.Context, in *SetupInput) (*SessionOutput, error) {
			s.mu.Lock()
			defer s.mu.Unlock()
			u, e := makeUser(domain.User{
				Username: in.Body.Username,
				Name:     in.Body.Username,
				Role:     domain.RoleAdmin,
				Locale:   "zh-CN",
				Timezone: in.Body.Timezone,
				Enabled:  true,
			})
			if e != nil {
				return nil, huma.Error422UnprocessableEntity(e.Error())
			}
			hash, e := security.HashPassword(in.Body.Password)
			if e != nil {
				return nil, huma.Error422UnprocessableEntity(e.Error())
			}
			var result *SessionOutput
			e = s.Store.WithTx(ctx, func(t *store.Tx) error {
				users, e := t.List(ctx, "users")
				if e != nil {
					return e
				}
				if len(users) != 0 {
					return huma.Error409Conflict("Setup already completed")
				}

				if e = t.Put(
					ctx,
					"users",
					u.ID,
					domain.UserRecord{User: u, PasswordHash: hash},
				); e != nil {
					return e
				}
				settings := domain.DefaultSettings()
				if in.Body.OrganizationName != "" {
					settings.OrganizationName = in.Body.OrganizationName
				}
				settings.Timezone = u.Timezone
				if e = t.Put(
					ctx,
					"settings",
					"organization",
					settings,
				); e != nil {
					return e
				}
				if e = audit(
					context.WithValue(ctx, contextKey{}, identity{User: u}),
					t,
					"setup",
					"users",
					u.ID,
				); e != nil {
					return e
				}
				result, e = s.createSession(ctx, t, u)
				return e
			})
			if e != nil {
				if se, ok := e.(huma.StatusError); ok {
					return nil, se
				}
				return nil, apiError(ctx, e)
			}
			return result, nil
		},
	)
	huma.Register(
		s.API,
		huma.Operation{
			OperationID: "getSession",
			Method:      "GET",
			Path:        "/api/v1/session",
		},
		func(ctx context.Context, _ *struct{}) (*Output[SessionBody], error) {
			id := currentIdentity(ctx)
			var body SessionBody
			if id.User.ID != "" {
				body.User = &id.User
				body.CSRFToken = id.Session.CSRFToken
			}
			return &Output[SessionBody]{Body: body}, nil
		},
	)
	huma.Register(
		s.API,
		huma.Operation{
			OperationID: "createSession",
			Method:      "POST",
			Path:        "/api/v1/session",
		},
		func(ctx context.Context, in *CreateInput[Credentials]) (*SessionOutput, error) {
			users, e := list[domain.UserRecord](ctx, s.Store, "users")
			if e != nil {
				return nil, apiError(ctx, e)
			}
			var found domain.UserRecord
			for _, u := range users {
				if u.Username == strings.ToLower(strings.TrimSpace(in.Body.Username)) {
					found = u
					break
				}
			}
			hash := found.PasswordHash
			if hash == "" {
				hash = s.dummyPassword
			}
			if !security.CheckPassword(hash, in.Body.Password) || !found.Enabled {
				return nil, huma.Error401Unauthorized("Invalid credentials")
			}
			var out *SessionOutput
			e = s.Store.WithTx(
				ctx,
				func(t *store.Tx) error {
					var err error
					out, err = s.createSession(ctx, t, found.User)
					return err
				},
			)
			return out, apiError(ctx, e)
		},
	)
	huma.Register(
		s.API,
		huma.Operation{
			OperationID: "deleteSession",
			Method:      "DELETE",
			Path:        "/api/v1/session",
		},
		func(ctx context.Context, _ *struct{}) (*SessionOutput, error) {
			id := currentIdentity(ctx)
			if e := s.Store.Delete(ctx, "sessions", id.Session.ID); e != nil {
				return nil, apiError(ctx, e)
			}
			return &SessionOutput{
				SetCookie: (&http.Cookie{
					Name:     "octopulse_session",
					Path:     "/api/v1",
					MaxAge:   -1,
					HttpOnly: true,
					Secure:   s.Config.CookieSecure,
					SameSite: http.SameSiteStrictMode,
				}).String(),
			}, nil
		},
	)
	huma.Register(
		s.API,
		huma.Operation{
			OperationID: "listUsers",
			Method:      "GET",
			Path:        "/api/v1/users",
		},
		func(ctx context.Context, _ *struct{}) (*Output[Items[domain.User]], error) {
			if CurrentUser(ctx).Role != domain.RoleAdmin {
				return nil, huma.Error403Forbidden("Administrator permission required")
			}
			records, e := list[domain.UserRecord](ctx, s.Store, "users")
			if e != nil {
				return nil, apiError(ctx, e)
			}
			users := []domain.User{}
			for _, u := range records {
				users = append(users, u.User)
			}
			return &Output[Items[domain.User]]{Body: Items[domain.User]{Items: users}}, nil
		},
	)
	huma.Register(
		s.API,
		huma.Operation{
			OperationID: "createUser",
			Method:      "POST",
			Path:        "/api/v1/users",
		},
		func(ctx context.Context, in *CreateInput[UserWrite]) (*Output[domain.User], error) {
			return s.saveUser(ctx, "", in.Body)
		},
	)
	huma.Register(
		s.API,
		huma.Operation{
			OperationID: "updateUser",
			Method:      "PATCH",
			Path:        "/api/v1/users/{id}",
		},
		func(ctx context.Context, in *WriteInput[UserWrite]) (*Output[domain.User], error) {
			return s.saveUser(ctx, in.ID, in.Body)
		},
	)
	huma.Register(
		s.API,
		huma.Operation{
			OperationID: "deleteUser",
			Method:      "DELETE",
			Path:        "/api/v1/users/{id}",
		},
		func(ctx context.Context, in *IDInput) (*Output[Ack], error) {
			s.mu.Lock()
			defer s.mu.Unlock()
			e := s.Store.WithTx(ctx, func(t *store.Tx) error {
				var old domain.UserRecord
				if e := t.Get(
					ctx,
					"users",
					in.ID,
					&old,
				); e != nil {
					return e
				}
				if old.Role == domain.RoleAdmin && old.Enabled {
					records, e := t.List(ctx, "users")
					if e != nil {
						return e
					}
					var count int
					for _, r := range records {
						var u domain.UserRecord
						if e = jsonUser(r, &u); e != nil {
							return e
						}
						if u.Role == domain.RoleAdmin && u.Enabled {
							count++
						}
					}
					if count <= 1 {
						return huma.Error409Conflict("Cannot delete the last enabled administrator")
					}
				}
				if e := t.Delete(ctx, "users", in.ID); e != nil {
					return e
				}
				return audit(
					ctx,
					t,
					"delete",
					"users",
					in.ID,
				)
			})
			if e != nil {
				return nil, statusOrAPIError(ctx, e)
			}
			return &Output[Ack]{Body: Ack{OK: true}}, nil
		},
	)
}

func makeUser(u domain.User) (domain.User, error) {
	u.Username = strings.ToLower(strings.TrimSpace(u.Username))
	validUsernameLength := len(u.Username) > 0 && len(u.Username) <= 100
	if !validUsernameLength || strings.ContainsAny(u.Username, " \t\n/") {
		return domain.User{}, errors.New("invalid username")
	}
	if u.Role == "" {
		u.Role = domain.RoleOperator
	}
	switch u.Role {
	case domain.RoleAdmin, domain.RoleOperator, domain.RoleViewer:
	default:
		return domain.User{}, errors.New("invalid role")
	}
	if u.Locale == "" {
		u.Locale = "zh-CN"
	}
	if u.Locale != "zh-CN" && u.Locale != "en" {
		return domain.User{}, errors.New("invalid locale")
	}
	if u.Timezone == "" {
		u.Timezone = "UTC"
	}
	if _, e := time.LoadLocation(u.Timezone); e != nil {
		return domain.User{}, errors.New("invalid timezone")
	}
	u.ID = domain.ID()
	u.CreatedAt = domain.Now()
	u.UpdatedAt = domain.Now()
	return u, nil
}
func (s *Server) saveUser(ctx context.Context, id string, in UserWrite) (*Output[domain.User], error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	enabled := true
	if in.Enabled != nil {
		enabled = *in.Enabled
	}
	u, e := makeUser(domain.User{
		Username: in.Username,
		Name:     in.Name,
		Role:     in.Role,
		Locale:   in.Locale,
		Timezone: in.Timezone,
		Enabled:  enabled,
	})
	if e != nil {
		return nil, huma.Error422UnprocessableEntity(e.Error())
	}
	var old domain.UserRecord
	if id != "" {
		if e = s.Store.Get(
			ctx,
			"users",
			id,
			&old,
		); e != nil {
			return nil, apiError(ctx, e)
		}
		u.ID = id
		u.CreatedAt = old.CreatedAt
	}
	hash := old.PasswordHash
	if in.Password != "" || id == "" {
		hash, e = security.HashPassword(in.Password)
		if e != nil {
			return nil, huma.Error422UnprocessableEntity(e.Error())
		}
	}
	e = s.Store.WithTx(ctx, func(t *store.Tx) error {
		records, e := t.List(ctx, "users")
		if e != nil {
			return e
		}
		var admins int
		for _, r := range records {
			var record domain.UserRecord
			if e = jsonUser(r, &record); e != nil {
				return e
			}
			if record.Username == u.Username && record.ID != u.ID {
				return huma.Error409Conflict("Username already exists")
			}
			if record.ID != u.ID && record.Role == domain.RoleAdmin && record.Enabled {
				admins++
			}
		}
		if old.Role == domain.RoleAdmin && old.Enabled && (!u.Enabled || u.Role != domain.RoleAdmin) && admins == 0 {
			return huma.Error409Conflict("Cannot disable the last administrator")
		}
		if id != "" && in.Password != "" {
			sessions, err := t.List(ctx, "sessions")
			if err != nil {
				return err
			}
			for _, b := range sessions {
				var session domain.Session
				if err := decode(b, &session); err != nil {
					return err
				}
				if session.UserID == u.ID {
					if err := t.Delete(ctx, "sessions", session.ID); err != nil {
						return err
					}
				}
			}
		}
		if e = t.Put(
			ctx,
			"users",
			u.ID,
			domain.UserRecord{User: u, PasswordHash: hash},
		); e != nil {
			return e
		}
		return audit(
			ctx,
			t,
			"save",
			"users",
			u.ID,
		)
	})
	if e != nil {
		return nil, statusOrAPIError(ctx, e)
	}
	return &Output[domain.User]{Body: u}, nil
}
func statusOrAPIError(ctx context.Context, e error) error {
	if v, ok := e.(huma.StatusError); ok {
		return v
	}
	return apiError(ctx, e)
}
func jsonUser(b []byte, u *domain.UserRecord) error { return json.Unmarshal(b, u) }
func decode(b []byte, out any) error                { return json.Unmarshal(b, out) }
