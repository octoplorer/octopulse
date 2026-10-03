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

// maintenanceSpans decodes each policy once, regardless of how many monitors
// or aggregate buckets are calculated from the same immutable read snapshot.
func maintenanceSpans(raw []json.RawMessage) map[string][]span {
	result := map[string][]span{}
	for _, row := range raw {
		var maintenance domain.Maintenance
		if json.Unmarshal(row, &maintenance) != nil {
			continue
		}
		seen := map[string]bool{}
		for _, id := range maintenance.MonitorIDs {
			if !seen[id] {
				result[id] = append(result[id], span{maintenance.StartsAt, maintenance.EndsAt})
				seen[id] = true
			}
		}
	}
	for id, spans := range result {
		result[id] = union(spans)
	}
	return result
}

type availabilityInput struct {
	created                int64
	intervals              []store.Interval
	intervalEnds           []int64 // Prefix maximum permits binary search even for overlapping intervals.
	maintenance            []span
	collectionThrough      int64
	hasCollectionWatermark bool
}

func prepareAvailability(snapshot store.StatisticsSnapshot, maintenance []span) availabilityInput {
	var monitor domain.Monitor
	_ = json.Unmarshal(snapshot.Monitor.ConfigJSON, &monitor)
	intervals := append([]store.Interval(nil), snapshot.Intervals...)
	sort.SliceStable(intervals, func(i, j int) bool { return intervals[i].StartedAt < intervals[j].StartedAt })
	ends := make([]int64, len(intervals))
	var end int64
	for i, interval := range intervals {
		current := int64(1<<63 - 1)
		if interval.EndedAt != nil {
			current = *interval.EndedAt
		}
		end = max(end, current)
		ends[i] = end
	}
	return availabilityInput{created: monitor.CreatedAt, intervals: intervals, intervalEnds: ends, maintenance: maintenance, collectionThrough: snapshot.CollectionThrough, hasCollectionWatermark: snapshot.HasCollectionWatermark}
}

func calculate(snapshot store.StatisticsSnapshot, from, to int64) domain.Availability {
	return prepareAvailability(snapshot, maintenanceSpans(snapshot.Maintenance)[snapshot.Monitor.ID]).calculate(from, to)
}

func (input availabilityInput) calculate(from, to int64) domain.Availability {
	result := domain.Availability{From: from, To: to}
	created := max(input.created, from)
	excluded := []span{}
	start := sort.Search(len(input.maintenance), func(i int) bool { return input.maintenance[i].to > created })
	for _, block := range input.maintenance[start:] {
		if block.from >= to {
			break
		}
		if interval, ok := clip(block, created, to); ok {
			excluded = append(excluded, interval)
		}
	}
	// A backfill can span thousands of buckets. Search the snapshot once per
	// bucket instead of rescanning every historical state interval each time.
	first := sort.Search(len(input.intervals), func(i int) bool { return input.intervalEnds[i] > created })
	last := sort.Search(len(input.intervals), func(i int) bool { return input.intervals[i].StartedAt >= to })
	if first > last {
		first = last
	}
	intervals := input.intervals[first:last]
	for _, interval := range intervals {
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
	for _, interval := range intervals {
		if interval.State != domain.StateUp && interval.State != domain.StateDown {
			continue
		}
		end := to
		if interval.EndedAt != nil {
			end = *interval.EndedAt
		} else if input.hasCollectionWatermark && input.collectionThrough < end {
			end = input.collectionThrough
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
