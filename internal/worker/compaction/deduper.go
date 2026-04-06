package compaction

func dedupe(events []Event) (deduped []Event, removed int) {

	seen := make(map[string]struct{}, len(events))

	deduped = make([]Event, 0, len(events))

	for i := range events {

		if _, ok := seen[events[i].ID]; ok {
			removed++
			continue
		}
		seen[events[i].ID] = struct{}{}
		deduped = append(deduped, events[i])
	}
	return deduped, removed
}

func groupByDay(events []Event) map[string][]Event {
	groups := make(map[string][]Event)
	for i := range events {
		day := events[i].OccurredDay()
		groups[day] = append(groups[day], events[i])
	}
	return groups
}
