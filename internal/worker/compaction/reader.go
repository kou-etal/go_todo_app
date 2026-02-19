package compaction

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// s3から読む部分
// readRawEvents は指定された claim の範囲の raw .jsonl ファイルを全て読み込む。
// backfill window 分をバーする。
func readRawEvents(
	ctx context.Context,
	storage ObjectStorage,
	rawPrefix string,
	//これconfigから取得したらよくない->rawPrefixしか使わないならそれだけ持つ。コードが分かりやすくなる。Explicit Dependency
	//使うやつだけ与える。
	claimDates []time.Time, //これは複数。backfilから作る。2/15->2/13,2/14
	metrics *Metrics,
) ([]Event, error) {
	var allEvents []Event

	for _, date := range claimDates {
		events, err := readEventsForDate(ctx, storage, rawPrefix, date, metrics)
		if err != nil {
			return nil, fmt.Errorf("read events for %s: %w", date.Format("2006-01-02"), err)
		}
		allEvents = append(allEvents, events...)
	}
	return allEvents, nil
}

// readEventsForDate は 1day(間パーティション）の .jsonl を読む。
func readEventsForDate(
	ctx context.Context,
	storage ObjectStorage,
	rawPrefix string,
	date time.Time,
	metrics *Metrics,
) ([]Event, error) {
	var events []Event

	for hour := 0; hour < 24; hour++ {
		prefix := fmt.Sprintf("%s/year=%04d/month=%02d/day=%02d/hour=%02d/",
			rawPrefix, date.Year(), date.Month(), date.Day(), hour,
		) //hourで区切ってるから取得もhour単位。

		listStart := time.Now()
		keys, err := storage.List(ctx, prefix) //Listはkey返す。データは返さない責務。
		listDur := time.Since(listStart).Seconds()
		metrics.S3ListDuration.Add(listDur)
		metrics.S3ListCalls.Add(1)
		metrics.S3ObjectsListed.Add(float64(len(keys)))
		if err != nil {
			return nil, fmt.Errorf("list prefix=%s: %w", prefix, err)
		}

		for _, key := range keys {
			if !strings.HasSuffix(key, ".jsonl") {
				continue
			}
			readStart := time.Now()
			parsed, err := readJSONLFile(ctx, storage, key)
			metrics.S3ReadDuration.Add(time.Since(readStart).Seconds())
			if err != nil {
				return nil, fmt.Errorf("read jsonl key=%s: %w", key, err)
			}
			events = append(events, parsed...)
		}
	}
	return events, nil
}

// readJSONLFile は S3 上の .jsonl ファイルを読み、Event スライスにパースする。
func readJSONLFile(ctx context.Context, storage ObjectStorage, key string) ([]Event, error) {
	body, err := storage.Get(ctx, key) //これがkeyからデータ受け取る。
	if err != nil {
		return nil, err
	}
	defer body.Close() //bodyはデータではなくio.Reader

	var events []Event
	scanner := bufio.NewScanner(body) //標準記法
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var e Event
		if err := json.Unmarshal(line, &e); err != nil {
			return nil, fmt.Errorf("unmarshal line: %w", err)
		}
		events = append(events, e)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan jsonl: %w", err)
	}
	return events, nil
}

// claimDatesForTarget は target から backfill window 分の claim リストを生成する。
// これで受け取ったdaysを代入する。
func claimDatesForTarget(target time.Time, backfillWindow int) []time.Time {
	dates := make([]time.Time, backfillWindow)
	for i := range backfillWindow {
		//r i := 0; i < backfillWindow; i++ と同様。新しい記法。
		//AddDate(year, month, day)。i days ago
		dates[i] = target.AddDate(0, 0, -i)
	}
	return dates
}
