package conflicts

import (
	"sort"
	"time"

	"github.com/yjwong/lark-cli/internal/api"
)

// EventTimeSlot represents parsed time boundaries for an event
type EventTimeSlot struct {
	ID     string
	Start  time.Time
	End    time.Time
	AllDay bool
}

// Options configures conflict detection
type Options struct {
	BufferMinutes int
}

// Result contains conflict detection results
type Result struct {
	ConflictMap  map[string][]string // eventID -> conflicting IDs
	Conflicts    []api.Conflict
	HasConflicts bool
}

// ParseEventTimes converts OutputEvents to EventTimeSlots for detection
func ParseEventTimes(events []api.OutputEvent, loc *time.Location) ([]EventTimeSlot, error) {
	slots := make([]EventTimeSlot, 0, len(events))

	for _, e := range events {
		slot := EventTimeSlot{
			ID:     e.ID,
			AllDay: e.AllDay,
		}

		if e.AllDay {
			// All-day events use YYYY-MM-DD format
			start, err := time.ParseInLocation("2006-01-02", e.Start, loc)
			if err != nil {
				return nil, err
			}
			end, err := time.ParseInLocation("2006-01-02", e.End, loc)
			if err != nil {
				return nil, err
			}
			// All-day events: start at midnight, end is exclusive (so end date is actually end of previous day)
			slot.Start = start
			slot.End = end // Keep as-is; comparison logic will handle exclusivity
		} else {
			// Timed events use RFC3339
			start, err := time.Parse(time.RFC3339, e.Start)
			if err != nil {
				return nil, err
			}
			end, err := time.Parse(time.RFC3339, e.End)
			if err != nil {
				return nil, err
			}
			slot.Start = start.In(loc)
			slot.End = end.In(loc)
		}

		slots = append(slots, slot)
	}

	return slots, nil
}

// Detect finds all conflicts in the given events
func Detect(slots []EventTimeSlot, opts Options) Result {
	result := Result{
		ConflictMap: make(map[string][]string),
		Conflicts:   []api.Conflict{},
	}

	if len(slots) < 2 {
		return result
	}

	// Sort by start time
	sort.Slice(slots, func(i, j int) bool {
		if slots[i].Start.Equal(slots[j].Start) {
			return slots[i].End.Before(slots[j].End)
		}
		return slots[i].Start.Before(slots[j].Start)
	})

	// Compare each pair
	for i := 0; i < len(slots); i++ {
		for j := i + 1; j < len(slots); j++ {
			a, b := slots[i], slots[j]

			// Optimization: if b starts after a ends + buffer, no more conflicts for a
			bufferDuration := time.Duration(opts.BufferMinutes) * time.Minute
			if b.Start.After(a.End.Add(bufferDuration)) || b.Start.Equal(a.End.Add(bufferDuration)) {
				break
			}

			// Check for overlap: a.Start < b.End AND b.Start < a.End
			// Note: events ending exactly when another starts are NOT conflicts
			if a.Start.Before(b.End) && b.Start.Before(a.End) {
				// Calculate overlap duration
				overlapStart := a.Start
				if b.Start.After(a.Start) {
					overlapStart = b.Start
				}
				overlapEnd := a.End
				if b.End.Before(a.End) {
					overlapEnd = b.End
				}
				overlapMinutes := int(overlapEnd.Sub(overlapStart).Minutes())

				result.Conflicts = append(result.Conflicts, api.Conflict{
					Type:           "overlap",
					EventIDs:       []string{a.ID, b.ID},
					OverlapMinutes: overlapMinutes,
				})
				addToConflictMap(result.ConflictMap, a.ID, b.ID)
			} else if opts.BufferMinutes > 0 {
				// Check for insufficient buffer (no overlap, but gap < buffer)
				gap := b.Start.Sub(a.End)
				if gap >= 0 && gap < bufferDuration {
					result.Conflicts = append(result.Conflicts, api.Conflict{
						Type:                  "insufficient_buffer",
						EventIDs:              []string{a.ID, b.ID},
						GapMinutes:            int(gap.Minutes()),
						RequiredBufferMinutes: opts.BufferMinutes,
					})
					addToConflictMap(result.ConflictMap, a.ID, b.ID)
				}
			}
		}
	}

	result.HasConflicts = len(result.Conflicts) > 0
	return result
}

// ApplyToEvents adds conflict information to OutputEvents
func ApplyToEvents(events []api.OutputEvent, result Result) {
	for i := range events {
		if conflicts, ok := result.ConflictMap[events[i].ID]; ok {
			events[i].ConflictsWith = conflicts
		}
	}
}

func addToConflictMap(m map[string][]string, a, b string) {
	m[a] = append(m[a], b)
	m[b] = append(m[b], a)
}
