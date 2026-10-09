package server

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/octoplorer/octopulse/internal/domain"
	"github.com/octoplorer/octopulse/internal/notify"
	"github.com/octoplorer/octopulse/internal/store"
)

type resourcePointer[T any] interface {
	*T
	SetID(string)
}
type resourceHooks[T any] struct {
	Save    func(context.Context, *store.Tx, *T, *T) error
	Delete  func(context.Context, *store.Tx, string) error
	Project func(context.Context, T) T
	Changed func()
}

func collection[T any, P resourcePointer[T]](s *Server, kind string, h resourceHooks[T]) {
	path := "/api/v1/" + kind
	huma.Register(
		s.API,
		huma.Operation{
			OperationID: "list" + strings.Title(kind),
			Method:      "GET",
			Path:        path,
		},
		func(ctx context.Context, _ *struct{}) (*Output[Items[T]], error) {
			items, e := list[T](ctx, s.Store, kind)
			if e != nil {
				return nil, apiError(ctx, e)
			}
			if h.Project != nil {
				for i := range items {
					items[i] = h.Project(ctx, items[i])
				}
			}
			return &Output[Items[T]]{Body: Items[T]{Items: items}}, nil
		},
	)
	huma.Register(
		s.API,
		huma.Operation{
			OperationID: "get" + strings.Title(kind),
			Method:      "GET",
			Path:        path + "/{id}",
		},
		func(ctx context.Context, in *IDInput) (*Output[T], error) {
			var v T
			if e := s.Store.Get(
				ctx,
				kind,
				in.ID,
				&v,
			); e != nil {
				return nil, apiError(ctx, e)
			}
			if h.Project != nil {
				v = h.Project(ctx, v)
			}
			return &Output[T]{Body: v}, nil
		},
	)
	save := func(ctx context.Context, id string, v T) (*Output[T], error) {
		s.mu.Lock()
		defer s.mu.Unlock()
		if id == "" {
			id = domain.ID()
		}
		P(&v).SetID(id)
		var old *T
		e := s.Store.WithTx(ctx, func(t *store.Tx) error {
			var o T
			e := t.Get(
				ctx,
				kind,
				id,
				&o,
			)
			if e == nil {
				old = &o
			} else if !errors.Is(e, store.ErrNotFound) {
				return e
			}
			if h.Save != nil {
				if e = h.Save(
					ctx,
					t,
					&v,
					old,
				); e != nil {
					return e
				}
			}
			if e = t.Put(
				ctx,
				kind,
				id,
				v,
			); e != nil {
				return e
			}
			return audit(
				ctx,
				t,
				"save",
				kind,
				id,
			)
		})
		if e != nil {
			return nil, statusOrAPIError(ctx, e)
		}
		if h.Project != nil {
			v = h.Project(ctx, v)
		}
		if h.Changed != nil {
			h.Changed()
		}
		return &Output[T]{Body: v}, nil
	}
	huma.Register(
		s.API,
		huma.Operation{
			OperationID: "create" + strings.Title(kind),
			Method:      "POST",
			Path:        path,
		},
		func(ctx context.Context, in *CreateInput[T]) (*Output[T], error) {
			return save(ctx, "", in.Body)
		},
	)
	huma.Register(
		s.API,
		huma.Operation{
			OperationID: "update" + strings.Title(kind),
			Method:      "PATCH",
			Path:        path + "/{id}",
		},
		func(ctx context.Context, in *WriteInput[T]) (*Output[T], error) {
			var exists T
			if e := s.Store.Get(
				ctx,
				kind,
				in.ID,
				&exists,
			); e != nil {
				return nil, apiError(ctx, e)
			}
			return save(ctx, in.ID, in.Body)
		},
	)
	huma.Register(
		s.API,
		huma.Operation{
			OperationID: "delete" + strings.Title(kind),
			Method:      "DELETE",
			Path:        path + "/{id}",
		},
		func(ctx context.Context, in *IDInput) (*Output[Ack], error) {
			s.mu.Lock()
			defer s.mu.Unlock()
			e := s.Store.WithTx(ctx, func(t *store.Tx) error {
				if h.Delete != nil {
					if e := h.Delete(ctx, t, in.ID); e != nil {
						return e
					}
				}
				if e := t.Delete(ctx, kind, in.ID); e != nil {
					return e
				}
				return audit(
					ctx,
					t,
					"delete",
					kind,
					in.ID,
				)
			})
			if e == nil && h.Changed != nil {
				h.Changed()
			}
			return &Output[Ack]{Body: Ack{OK: true}}, statusOrAPIError(ctx, e)
		},
	)
}

func (s *Server) registerConfiguration() {
	collection[domain.Channel, *domain.Channel](
		s,
		"channels",
		resourceHooks[domain.Channel]{Save: func(ctx context.Context, t *store.Tx, v, old *domain.Channel) error {
			if strings.TrimSpace(v.Name) == "" {
				return huma.Error422UnprocessableEntity("Channel name is required")
			}
			var secret domain.SecretRecord
			if e := t.Get(
				ctx,
				"secrets",
				v.ServiceURLSecretID,
				&secret,
			); e != nil {
				return huma.Error422UnprocessableEntity("Notification URL secret is unavailable")
			}
			raw, e := s.Vault.Decrypt(secret.ID, secret.Ciphertext)
			if e != nil || notify.ValidateURL(raw) != nil {
				return huma.Error422UnprocessableEntity("Secret must contain a valid Shoutrrr service URL")
			}
			v.CreatedAt = domain.Now()
			if old != nil {
				v.CreatedAt = old.CreatedAt
			}
			v.UpdatedAt = domain.Now()
			return nil
		}, Delete: func(ctx context.Context, t *store.Tx, id string) error {
			monitors, e := t.List(ctx, "monitors")
			if e != nil {
				return e
			}
			for _, b := range monitors {
				var m domain.Monitor
				if e = decode(b, &m); e != nil {
					return e
				}
				if hasID(m.NotificationChannelIDs, id) {
					return huma.Error409Conflict("Channel is referenced by a monitor")
				}
			}
			return nil
		}},
	)
	collection[domain.Maintenance, *domain.Maintenance](
		s,
		"maintenance",
		resourceHooks[domain.Maintenance]{Save: func(ctx context.Context, t *store.Tx, v, old *domain.Maintenance) error {
			if strings.TrimSpace(v.Name) == "" || v.EndsAt <= v.StartsAt {
				return huma.Error422UnprocessableEntity("Maintenance requires a name and valid time range")
			}
			if v.Timezone == "" {
				v.Timezone = "UTC"
			}
			if _, e := time.LoadLocation(v.Timezone); e != nil {
				return huma.Error422UnprocessableEntity("Invalid maintenance timezone")
			}
			for _, id := range v.MonitorIDs {
				if _, e := t.GetMonitor(ctx, id); e != nil {
					return huma.Error422UnprocessableEntity("Maintenance monitor is unavailable")
				}
			}
			for _, id := range v.PageIDs {
				var p domain.Page
				if e := t.Get(
					ctx,
					"pages",
					id,
					&p,
				); e != nil {
					return huma.Error422UnprocessableEntity("Maintenance page is unavailable")
				}
			}
			v.CreatedAt = domain.Now()
			if old != nil {
				v.CreatedAt = old.CreatedAt
			}
			v.UpdatedAt = domain.Now()
			return nil
		}, Changed: func() {
			if s.Wake != nil {
				s.Wake()
			}
		}},
	)
	collection[domain.Incident, *domain.Incident](
		s,
		"incidents",
		resourceHooks[domain.Incident]{Save: func(ctx context.Context, t *store.Tx, v, old *domain.Incident) error {
			if strings.TrimSpace(v.Title) == "" {
				return huma.Error422UnprocessableEntity("Incident title is required")
			}
			if v.Status == "" {
				v.Status = "investigating"
			}
			if v.Impact == "" {
				v.Impact = "none"
			}
			v.CreatedAt = domain.Now()
			v.Updates = []domain.IncidentUpdate{}
			for _, id := range v.PageIDs {
				var p domain.Page
				if e := t.Get(
					ctx,
					"pages",
					id,
					&p,
				); e != nil {
					return huma.Error422UnprocessableEntity("Incident page is unavailable")
				}
			}
			for _, id := range v.MonitorIDs {
				if _, e := t.GetMonitor(ctx, id); e != nil {
					return huma.Error422UnprocessableEntity("Incident monitor is unavailable")
				}
			}
			if old != nil {
				v.CreatedAt = old.CreatedAt
				v.Updates = old.Updates
			}
			v.UpdatedAt = domain.Now()
			v.ResolvedAt = 0
			if v.Status == "resolved" {
				v.ResolvedAt = domain.Now()
			}
			return nil
		}},
	)
	s.registerSecrets()
	huma.Register(
		s.API,
		huma.Operation{
			OperationID: "getSettings",
			Method:      "GET",
			Path:        "/api/v1/settings",
		},
		func(ctx context.Context, _ *struct{}) (*Output[domain.Settings], error) {
			v := domain.DefaultSettings()
			e := s.Store.Get(
				ctx,
				"settings",
				"organization",
				&v,
			)
			if e != nil && !errors.Is(e, store.ErrNotFound) {
				return nil, apiError(ctx, e)
			}
			return &Output[domain.Settings]{Body: v}, nil
		},
	)
	huma.Register(
		s.API,
		huma.Operation{
			OperationID: "updateSettings",
			Method:      "PATCH",
			Path:        "/api/v1/settings",
		},
		func(ctx context.Context, in *CreateInput[domain.Settings]) (*Output[domain.Settings], error) {
			s.mu.Lock()
			defer s.mu.Unlock()
			v := in.Body
			if strings.TrimSpace(v.OrganizationName) == "" {
				return nil, huma.Error422UnprocessableEntity("Organization name is required")
			}
			if _, e := time.LoadLocation(v.Timezone); e != nil {
				return nil, huma.Error422UnprocessableEntity("Invalid timezone")
			}
			for _, h := range v.AllowedDomains {
				invalidHostname := h == "" || hostname(h) != h || strings.ContainsAny(h, "/:?# ")
				if invalidHostname || s.adminHost(h) {
					return nil, huma.Error422UnprocessableEntity(
						"Allowed domain must be a public hostname separate from management hosts",
					)
				}
			}
			e := s.Store.WithTx(ctx, func(t *store.Tx) error {
				pages, e := t.List(ctx, "pages")
				if e != nil {
					return e
				}
				for _, b := range pages {
					var p domain.Page
					if e := decode(b, &p); e != nil {
						return e
					}
					domainRemoved := p.Domain != "" && !hasID(v.AllowedDomains, p.Domain)
					if !domainRemoved && p.PublishedDomain != "" {
						domainRemoved = !hasID(v.AllowedDomains, p.PublishedDomain)
					}
					if domainRemoved {
						return huma.Error409Conflict("Unbind status pages before removing their domains")
					}
				}
				if e := t.Put(
					ctx,
					"settings",
					"organization",
					v,
				); e != nil {
					return e
				}
				return audit(
					ctx,
					t,
					"save",
					"settings",
					"organization",
				)
			})
			return &Output[domain.Settings]{Body: v}, apiError(ctx, e)
		},
	)
	huma.Register(
		s.API,
		huma.Operation{
			OperationID: "listAudit",
			Method:      "GET",
			Path:        "/api/v1/audit",
		},
		func(ctx context.Context, _ *struct{}) (*Output[Items[domain.Audit]], error) {
			v, e := list[domain.Audit](ctx, s.Store, "audit")
			return &Output[Items[domain.Audit]]{Body: Items[domain.Audit]{Items: v}}, apiError(ctx, e)
		},
	)
}

type SecretWrite struct {
	Name  string `json:"name"`
	Value string `json:"value" minLength:"1" maxLength:"2097152"`
}

func (s *Server) registerSecrets() {
	huma.Register(
		s.API,
		huma.Operation{
			OperationID: "listSecrets",
			Method:      "GET",
			Path:        "/api/v1/secrets",
		},
		func(ctx context.Context, _ *struct{}) (*Output[Items[domain.Secret]], error) {
			v, e := list[domain.SecretRecord](ctx, s.Store, "secrets")
			if e != nil {
				return nil, apiError(ctx, e)
			}
			r := []domain.Secret{}
			for _, secret := range v {
				r = append(r, secret.Secret)
			}
			return &Output[Items[domain.Secret]]{Body: Items[domain.Secret]{Items: r}}, nil
		},
	)
	save := func(ctx context.Context, id string, v SecretWrite) (*Output[domain.Secret], error) {
		s.mu.Lock()
		defer s.mu.Unlock()
		if strings.TrimSpace(v.Name) == "" || v.Value == "" {
			return nil, huma.Error422UnprocessableEntity("Secret name and replacement value are required")
		}
		record := domain.SecretRecord{
			Secret: domain.Secret{
				ID:        id,
				Name:      v.Name,
				CreatedAt: domain.Now(),
				UpdatedAt: domain.Now(),
			},
		}
		if id == "" {
			record.ID = domain.ID()
		} else {
			var old domain.SecretRecord
			if e := s.Store.Get(
				ctx,
				"secrets",
				id,
				&old,
			); e != nil {
				return nil, apiError(ctx, e)
			}
			record.CreatedAt = old.CreatedAt
		}
		record.Ciphertext = s.Vault.Encrypt(record.ID, v.Value)
		e := s.Store.WithTx(ctx, func(t *store.Tx) error {
			if e := t.Put(
				ctx,
				"secrets",
				record.ID,
				record,
			); e != nil {
				return e
			}
			return audit(
				ctx,
				t,
				"save",
				"secrets",
				record.ID,
			)
		})
		return &Output[domain.Secret]{Body: record.Secret}, apiError(ctx, e)
	}
	huma.Register(
		s.API,
		huma.Operation{
			OperationID:  "createSecret",
			Method:       "POST",
			Path:         "/api/v1/secrets",
			MaxBodyBytes: 8 << 20,
		},
		func(ctx context.Context, in *CreateInput[SecretWrite]) (*Output[domain.Secret], error) {
			return save(ctx, "", in.Body)
		},
	)
	huma.Register(
		s.API,
		huma.Operation{
			OperationID:  "updateSecret",
			Method:       "PATCH",
			Path:         "/api/v1/secrets/{id}",
			MaxBodyBytes: 8 << 20,
		},
		func(ctx context.Context, in *WriteInput[SecretWrite]) (*Output[domain.Secret], error) {
			return save(ctx, in.ID, in.Body)
		},
	)
	huma.Register(
		s.API,
		huma.Operation{
			OperationID: "deleteSecret",
			Method:      "DELETE",
			Path:        "/api/v1/secrets/{id}",
		},
		func(ctx context.Context, in *IDInput) (*Output[Ack], error) {
			s.mu.Lock()
			defer s.mu.Unlock()
			e := s.Store.WithTx(ctx, func(t *store.Tx) error {
				for _, kind := range []string{"monitors", "channels"} {
					rows, e := t.List(ctx, kind)
					if e != nil {
						return e
					}
					for _, b := range rows {
						if kind == "monitors" {
							var m domain.Monitor
							if e = decode(b, &m); e != nil {
								return e
							}
							if hasID(m.SecretReferences(), in.ID) {
								return huma.Error409Conflict("Secret is referenced by a monitor")
							}
						} else {
							var c domain.Channel
							if e = decode(b, &c); e != nil {
								return e
							}
							if c.ServiceURLSecretID == in.ID {
								return huma.Error409Conflict("Secret is referenced by a notification channel")
							}
						}
					}
				}
				var beszel domain.BeszelConfig
				e := t.Get(
					ctx,
					"beszel",
					"config",
					&beszel,
				)
				if e == nil && beszel.PasswordSecretID == in.ID {
					return huma.Error409Conflict("Secret is referenced by Beszel")
				}
				if e != nil && !errors.Is(e, store.ErrNotFound) {
					return e
				}
				if e := t.Delete(ctx, "secrets", in.ID); e != nil {
					return e
				}
				return audit(
					ctx,
					t,
					"delete",
					"secrets",
					in.ID,
				)
			})
			return &Output[Ack]{Body: Ack{OK: true}}, apiError(ctx, e)
		},
	)
}
