// Package secrets resolves encrypted references independently of HTTP handlers.
package secrets

import (
	"context"
	"errors"

	"github.com/octoplorer/octopulse/internal/domain"
	"github.com/octoplorer/octopulse/internal/security"
	"github.com/octoplorer/octopulse/internal/store"
)

type Resolver struct {
	store *store.Store
	vault *security.Vault
}

func New(st *store.Store, vault *security.Vault) *Resolver {
	return &Resolver{store: st, vault: vault}
}

func (r *Resolver) ResolveSecret(ctx context.Context, id string) (string, error) {
	if r.store == nil || r.vault == nil {
		return "", errors.New("secret resolver is unavailable")
	}
	var record domain.SecretRecord
	if err := r.store.Get(
		ctx,
		"secrets",
		id,
		&record,
	); err != nil {
		return "", errors.New("secret reference is unavailable")
	}
	return r.vault.Decrypt(id, record.Ciphertext)
}
