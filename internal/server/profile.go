package server

import (
	"context"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/octoplorer/octopulse/internal/domain"
	"github.com/octoplorer/octopulse/internal/security"
	"github.com/octoplorer/octopulse/internal/store"
)

type ProfileWrite struct {
	Name        string `json:"name"`
	Locale      string `json:"locale" enum:"zh-CN,en"`
	Timezone    string `json:"timezone"`
	OldPassword string `json:"oldPassword,omitempty"`
	Password    string `json:"password,omitempty"`
}

func (s *Server) registerProfile() {
	huma.Register(s.API, huma.Operation{OperationID: "updateProfile", Method: "PATCH", Path: "/api/v1/profile"}, func(ctx context.Context, in *CreateInput[ProfileWrite]) (*Output[domain.User], error) {
		v := in.Body
		if _, e := time.LoadLocation(v.Timezone); e != nil {
			return nil, huma.Error422UnprocessableEntity("Invalid timezone")
		}
		var original domain.UserRecord
		var passwordHash string
		if v.Password != "" {
			if err := s.Store.Get(ctx, "users", CurrentUser(ctx).ID, &original); err != nil {
				return nil, apiError(ctx, err)
			}
			if !security.CheckPassword(original.PasswordHash, v.OldPassword) {
				return nil, huma.Error403Forbidden("Current password is incorrect")
			}
			var err error
			passwordHash, err = security.HashPassword(v.Password)
			if err != nil {
				return nil, huma.Error422UnprocessableEntity(err.Error())
			}
		}
		s.mu.Lock()
		defer s.mu.Unlock()
		var record domain.UserRecord
		e := s.Store.WithTx(ctx, func(t *store.Tx) error {
			if e := t.Get(ctx, "users", CurrentUser(ctx).ID, &record); e != nil {
				return e
			}
			if !record.Enabled {
				return huma.Error403Forbidden("Account is disabled")
			}
			if passwordHash != "" {
				if record.PasswordHash != original.PasswordHash {
					return huma.Error409Conflict("Password changed; retry with the current password")
				}
				record.PasswordHash = passwordHash
				sessions, e := t.List(ctx, "sessions")
				if e != nil {
					return e
				}
				for _, b := range sessions {
					var session domain.Session
					if e = decode(b, &session); e != nil {
						return e
					}
					if session.UserID == record.ID && session.ID != currentIdentity(ctx).Session.ID {
						if e = t.Delete(ctx, "sessions", session.ID); e != nil {
							return e
						}
					}
				}
			}
			record.Name = v.Name
			record.Locale = v.Locale
			record.Timezone = v.Timezone
			record.UpdatedAt = domain.Now()
			if e := t.Put(ctx, "users", record.ID, record); e != nil {
				return e
			}
			return audit(ctx, t, "profile", "users", record.ID)
		})
		return &Output[domain.User]{Body: record.User}, statusOrAPIError(ctx, e)
	})
}
