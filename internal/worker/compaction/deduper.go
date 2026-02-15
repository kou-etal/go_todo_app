package compaction

//compactionでは完全ではないけどevent_idで簡単に冪等。

// dedupe は event_id をキーに重複を除去する。最初に出現した行を保持。
func dedupe(events []Event) (deduped []Event, removed int) {
	//seenは種類判定。Goに set 型がないから map[key]struct{} を set として使うのが標準。
	//struct{} を値にするとメモリmin（値を持つ意味がない）
	seen := make(map[string]struct{}, len(events))
	//合計数はわからんけどmaxわかる場合はdeduped = make([]Event, 0, len(events))の記法
	deduped = make([]Event, 0, len(events))

	for i := range events {
		//値使わないからindexだけ回す。これは重いコピーの場合大事。
		if _, ok := seen[events[i].ID]; ok { //Goらしい記法
			removed++
			continue
		}
		seen[events[i].ID] = struct{}{}      //標準記法
		deduped = append(deduped, events[i]) //ここは値代入。使う値だけevents[i]で取得。
	}
	return deduped, removed
}

// groupByDay は events を occurred_at の日付でグルーピングする。
////compactionはoccured_atベース。
func groupByDay(events []Event) map[string][]Event { //mapの記法
	groups := make(map[string][]Event)
	for i := range events {
		day := events[i].OccurredDay()
		groups[day] = append(groups[day], events[i])
	}
	return groups
}
