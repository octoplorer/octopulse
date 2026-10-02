package statistics

import (
	"encoding/json"
	"sort"

	"github.com/octoplorer/octopulse/internal/domain"
	"github.com/octoplorer/octopulse/internal/store"
)

type span struct{ from, to int64 }

func clip(s span, from, to int64) (span, bool) {
	if s.from < from {
		s.from = from
	}
	if s.to > to {
		s.to = to
	}
	return s, s.from < s.to
}
func union(spans []span) []span {
	sort.Slice(spans, func(i, j int) bool {
		if spans[i].from == spans[j].from {
			return spans[i].to < spans[j].to
		}
		return spans[i].from < spans[j].from
	})
	result := []span{}
	for _, s := range spans {
		if s.from >= s.to {
			continue
		}
		if len(result) == 0 || s.from > result[len(result)-1].to {
			result = append(result, s)
		} else if s.to > result[len(result)-1].to {
			result[len(result)-1].to = s.to
		}
	}
	return result
}
func duration(spans []span) int64 {
	var total int64
	for _, s := range spans {
		total += s.to - s.from
	}
	return total
}
func subtract(s span, excluded []span) []span {
	result := []span{}
	cursor := s.from
	for _, block := range excluded {
		if block.to <= cursor {
			continue
		}
		if block.from >= s.to {
			break
		}
		if block.from > cursor {
			end := block.from
			if end > s.to {
				end = s.to
			}
			result = append(result, span{cursor, end})
		}
		if block.to > cursor {
			cursor = block.to
		}
		if cursor >= s.to {
			return result
		}
	}
	if cursor < s.to {
		result = append(result, span{cursor, s.to})
	}
	return result
}

func calculate(snapshot store.StatisticsSnapshot, from, to int64) domain.Availability {
	result := domain.Availability{From: from, To: to}
	var m domain.Monitor
	_ = json.Unmarshal(snapshot.Monitor.ConfigJSON, &m)
	created := m.CreatedAt
	if created < from {
		created = from
	}
	excluded := []span{}
	for _, row := range snapshot.Maintenance {
		var maintenance domain.Maintenance
		if json.Unmarshal(row, &maintenance) != nil {
			continue
		}
		for _, id := range maintenance.MonitorIDs {
			if id == snapshot.Monitor.ID {
				if interval, ok := clip(span{maintenance.StartsAt, maintenance.EndsAt}, created, to); ok {
					excluded = append(excluded, interval)
				}
				break
			}
		}
	}
	for _, interval := range snapshot.Intervals {
		if interval.State != "paused" {
			continue
		}
		end := to
		if interval.EndedAt != nil {
			end = *interval.EndedAt
		}
		if interval, ok := clip(span{interval.StartedAt, end}, created, to); ok {
			excluded = append(excluded, interval)
		}
	}
	excluded = union(excluded)
	result.ExcludedMs = duration(excluded)
	eligibleMS := to - from - result.ExcludedMs
	up, down := []span{}, []span{}
	for _, interval := range snapshot.Intervals {
		if interval.State != domain.StateUp && interval.State != domain.StateDown {
			continue
		}
		end := to
		if interval.EndedAt != nil {
			end = *interval.EndedAt
		} else if snapshot.HasCollectionWatermark && snapshot.CollectionThrough < end {
			end = snapshot.CollectionThrough
		}
		source, ok := clip(span{interval.StartedAt, end}, created, to)
		if !ok {
			continue
		}
		parts := subtract(source, excluded)
		if interval.State == domain.StateUp {
			up = append(up, parts...)
		} else {
			down = append(down, parts...)
		}
	}
	up = union(up)
	down = union(down)
	result.UpMs = duration(up)
	result.DownMs = duration(down)
	result.EffectiveMs = result.UpMs + result.DownMs
	result.UnknownMs = eligibleMS - result.EffectiveMs
	if result.UnknownMs < 0 {
		result.UnknownMs = 0
	} // Valid store intervals never overlap.
	if result.UpMs+result.DownMs > 0 {
		value := float64(result.UpMs) / float64(result.UpMs+result.DownMs) * 100
		result.Uptime = &value
	}
	if eligibleMS > 0 {
		value := float64(result.EffectiveMs) / float64(eligibleMS) * 100
		result.Coverage = &value
	}
	return result
}
