package server

import (
	"context"
	"errors"

	"github.com/danielgtaylor/huma/v2"
	"github.com/octoplorer/octopulse/internal/beszel"
	"github.com/octoplorer/octopulse/internal/domain"
	"github.com/octoplorer/octopulse/internal/store"
)

type BeszelHistoryInput struct {
	ID    string `path:"id"`
	Range string `query:"range" default:"24h" enum:"1h,12h,24h,1w,30d"`
}

func (s *Server) registerBeszel() *beszel.Client {
	client := beszel.New(s.Store, s.Secrets)
	huma.Register(
		s.API,
		huma.Operation{
			OperationID: "getBeszelConfig",
			Method:      "GET",
			Path:        "/api/v1/beszel/config",
		},
		func(ctx context.Context, _ *struct{}) (*Output[domain.BeszelConfig], error) {
			if CurrentUser(ctx).Role != domain.RoleAdmin {
				return nil, huma.Error403Forbidden("Administrator permission required")
			}
			cfg := domain.BeszelConfig{PollSeconds: 30}
			if err := s.Store.Get(
				ctx,
				"beszel",
				"config",
				&cfg,
			); err != nil && !errors.Is(err, store.ErrNotFound) {
				return nil, apiError(ctx, err)
			}
			return &Output[domain.BeszelConfig]{Body: cfg}, nil
		},
	)
	huma.Register(
		s.API,
		huma.Operation{
			OperationID: "updateBeszelConfig",
			Method:      "PATCH",
			Path:        "/api/v1/beszel/config",
		},
		func(ctx context.Context, in *CreateInput[domain.BeszelConfig]) (*Output[domain.BeszelConfig], error) {
			s.mu.Lock()
			defer s.mu.Unlock()
			if CurrentUser(ctx).Role != domain.RoleAdmin {
				return nil, huma.Error403Forbidden("Administrator permission required")
			}
			if err := beszel.ValidateConfig(&in.Body); err != nil {
				return nil, huma.Error422UnprocessableEntity(err.Error())
			}
			err := client.UpdateConfig(ctx, func() error {
				return s.Store.WithTx(ctx, func(tx *store.Tx) error {
					if in.Body.Enabled {
						var secret domain.SecretRecord
						if err := tx.Get(
							ctx,
							"secrets",
							in.Body.PasswordSecretID,
							&secret,
						); err != nil {
							return err
						}
					}
					if err := tx.Put(
						ctx,
						"beszel",
						"config",
						in.Body,
					); err != nil {
						return err
					}
					if err := tx.Delete(ctx, "beszelSnapshots", "systems"); err != nil && !errors.Is(err, store.ErrNotFound) {
						return err
					}
					return audit(
						ctx,
						tx,
						"configure",
						"beszel",
						"config",
					)
				})
			})
			if err != nil {
				return nil, apiError(ctx, err)
			}
			return &Output[domain.BeszelConfig]{Body: in.Body}, nil
		},
	)
	huma.Register(
		s.API,
		huma.Operation{
			OperationID: "listBeszelSystems",
			Method:      "GET",
			Path:        "/api/v1/beszel/systems",
		},
		func(ctx context.Context, _ *struct{}) (*Output[beszel.SystemsResponse], error) {
			result, err := client.Systems(ctx)
			if err != nil {
				return nil, apiError(ctx, err)
			}
			return &Output[beszel.SystemsResponse]{Body: result}, nil
		},
	)
	huma.Register(
		s.API,
		huma.Operation{
			OperationID: "getBeszelHistory",
			Method:      "GET",
			Path:        "/api/v1/beszel/systems/{id}/history",
		},
		func(ctx context.Context, in *BeszelHistoryInput) (*Output[beszel.HistoryResponse], error) {
			result, err := client.History(ctx, in.ID, in.Range)
			if err != nil {
				return nil, beszelError(err)
			}
			return &Output[beszel.HistoryResponse]{Body: result}, nil
		},
	)
	huma.Register(
		s.API,
		huma.Operation{
			OperationID: "listBeszelContainers",
			Method:      "GET",
			Path:        "/api/v1/beszel/systems/{id}/containers",
		},
		func(ctx context.Context, in *IDInput) (*Output[beszel.ContainersResponse], error) {
			result, err := client.Containers(ctx, in.ID)
			if err != nil {
				return nil, beszelError(err)
			}
			return &Output[beszel.ContainersResponse]{Body: result}, nil
		},
	)
	return client
}

func beszelError(err error) error {
	if errors.Is(err, beszel.ErrSchema) {
		return huma.Error422UnprocessableEntity("Invalid Beszel system or history range")
	}
	if errors.Is(err, beszel.ErrDisabled) {
		return huma.Error409Conflict(err.Error())
	}
	return huma.Error502BadGateway(err.Error())
}
