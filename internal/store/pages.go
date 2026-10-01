package store

import (
	"context"
	"fmt"
	"strings"
)

// BindPage reserves the normalized slug and hostnames inside the same
// transaction as the page document. Any clash rolls back the entire edit.
func (t *Tx) BindPage(ctx context.Context, b PageBinding) error {
	b.Slug = strings.ToLower(strings.TrimSpace(b.Slug))
	if b.PageID == "" || b.Slug == "" || strings.ContainsAny(b.Slug, "/?#.") {
		return fmt.Errorf("invalid page binding")
	}
	_, err := t.tx.ExecContext(ctx, t.s.sql(`INSERT INTO page_slugs(page_id,slug) VALUES(?,?) ON CONFLICT(page_id) DO UPDATE SET slug=excluded.slug`), b.PageID, b.Slug)
	if err != nil {
		return mapError(err)
	}
	if _, err = t.tx.ExecContext(ctx, t.s.sql(`DELETE FROM page_domains WHERE page_id=?`), b.PageID); err != nil {
		return mapError(err)
	}
	seen := map[string]bool{}
	for _, domain := range b.Domains {
		domain = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(domain), "."))
		if domain == "" || strings.ContainsAny(domain, "/:?# ") {
			return fmt.Errorf("invalid page hostname")
		}
		if seen[domain] {
			continue
		}
		seen[domain] = true
		_, err = t.tx.ExecContext(ctx, t.s.sql(`INSERT INTO page_domains(domain,page_id) VALUES(?,?)`), domain, b.PageID)
		if err != nil {
			return mapError(err)
		}
	}
	return nil
}

func (s *Store) BindPage(ctx context.Context, b PageBinding) error {
	return s.WithTx(ctx, func(t *Tx) error { return t.BindPage(ctx, b) })
}

func (s *Store) PageIDBySlug(ctx context.Context, slug string) (string, error) {
	var id string
	err := s.read.QueryRowContext(ctx, s.sql(`SELECT page_id FROM page_slugs WHERE slug=?`), strings.ToLower(slug)).Scan(&id)
	return id, mapError(err)
}

func (s *Store) PageIDByDomain(ctx context.Context, domain string) (string, error) {
	var id string
	domain = strings.ToLower(strings.TrimSuffix(domain, "."))
	err := s.read.QueryRowContext(ctx, s.sql(`SELECT page_id FROM page_domains WHERE domain=?`), domain).Scan(&id)
	return id, mapError(err)
}

func (s *Store) PageBinding(ctx context.Context, id string) (PageBinding, error) {
	b := PageBinding{PageID: id, Domains: []string{}}
	if err := s.read.QueryRowContext(ctx, s.sql(`SELECT slug FROM page_slugs WHERE page_id=?`), id).Scan(&b.Slug); err != nil {
		return b, mapError(err)
	}
	rows, err := s.read.QueryContext(ctx, s.sql(`SELECT domain FROM page_domains WHERE page_id=? ORDER BY domain`), id)
	if err != nil {
		return b, mapError(err)
	}
	defer rows.Close()
	for rows.Next() {
		var domain string
		if err = rows.Scan(&domain); err != nil {
			return b, err
		}
		b.Domains = append(b.Domains, domain)
	}
	return b, mapError(rows.Err())
}
