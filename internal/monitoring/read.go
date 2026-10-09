package monitoring

import (
	"context"
	"encoding/json"

	"github.com/octoplorer/octopulse/internal/domain"
	"github.com/octoplorer/octopulse/internal/store"
)

func (s *Service) project(snapshot store.MonitorSnapshot) (domain.Monitor, error) {
	row := snapshot.Monitor
	var m domain.Monitor
	if err := json.Unmarshal(row.ConfigJSON, &m); err != nil {
		return m, err
	}
	if rt := snapshot.Runtime; rt != nil && rt.ConfigVersion == row.ConfigVersion && rt.Generation == row.Generation {
		m.State = rt.State
		m.FailureCount = int(rt.Failures)
		m.SuccessCount = int(rt.Successes)
		m.LastCheckedAt = rt.LastCollectedAt
		if m.Heartbeat != nil {
			m.Heartbeat.LastReceivedAt = rt.HeartbeatAt
		}
	}
	var metadata struct {
		ConfigVersion        int64                     `json:"configVersion"`
		Certificate          *domain.CertificateConfig `json:"certificate"`
		HeartbeatSuccess     bool                      `json:"heartbeatSuccess"`
		HeartbeatDescription string                    `json:"heartbeatDescription"`
	}
	if len(snapshot.EngineJSON) > 0 {
		if err := json.Unmarshal(snapshot.EngineJSON, &metadata); err != nil {
			return m, err
		}
		if metadata.ConfigVersion == row.ConfigVersion {
			if m.Heartbeat != nil {
				m.Heartbeat.LastSuccess, m.Heartbeat.Description = metadata.HeartbeatSuccess, metadata.HeartbeatDescription
			}
			if m.Certificate != nil && metadata.Certificate != nil {
				m.Certificate.State = metadata.Certificate.State
				m.Certificate.ExpiresAt = metadata.Certificate.ExpiresAt
				m.Certificate.Fingerprint = metadata.Certificate.Fingerprint
				m.Certificate.DaysRemaining = metadata.Certificate.DaysRemaining
			}
		}
	}
	if s.NextCheck != nil {
		m.NextCheckAt = s.NextCheck(m.ID)
	}
	return m, nil
}

func (s *Service) Get(ctx context.Context, id string) (domain.Monitor, error) {
	items, err := s.Read(ctx, []string{id})
	if err != nil {
		return domain.Monitor{}, err
	}
	m, ok := items[id]
	if !ok {
		return m, store.ErrNotFound
	}
	return m, nil
}

func (s *Service) List(ctx context.Context) ([]domain.Monitor, error) {
	rows, err := s.Store.ListMonitorSnapshots(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]domain.Monitor, 0, len(rows))
	for _, row := range rows {
		m, err := s.project(row)
		if err != nil {
			return nil, err
		}
		items = append(items, m)
	}
	return items, nil
}

func (s *Service) Read(ctx context.Context, ids []string) (map[string]domain.Monitor, error) {
	rows, err := s.Store.ReadMonitorSnapshots(ctx, ids)
	if err != nil {
		return nil, err
	}
	items := make(map[string]domain.Monitor, len(rows))
	for id, row := range rows {
		m, err := s.project(row)
		if err != nil {
			return nil, err
		}
		items[id] = m
	}
	return items, nil
}
