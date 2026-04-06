package outbox

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

type jsonlRow struct {
	ID            string          `json:"id"`
	UserID        string          `json:"user_id"`
	TaskID        string          `json:"task_id"`
	RequestID     string          `json:"request_id"`
	EventType     string          `json:"event_type"`
	OccurredAt    time.Time       `json:"occurred_at"`
	SchemaVersion uint32          `json:"schema_version"`
	Payload       json.RawMessage `json:"payload"`
}

func buildJSONLines(records []ClaimedEvent) ([]byte, error) {

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)

	for i := range records {
		row := jsonlRow{
			ID:            records[i].ID,
			UserID:        records[i].UserID,
			TaskID:        records[i].TaskID,
			RequestID:     records[i].RequestID,
			EventType:     records[i].EventType,
			OccurredAt:    records[i].OccurredAt,
			SchemaVersion: records[i].SchemaVersion,
			Payload:       records[i].Payload,
		}
		if err := enc.Encode(row); err != nil {
			return nil, fmt.Errorf("encode jsonl row id=%s: %w", records[i].ID, err)
		}
	}

	return buf.Bytes(), nil
}

type manifest struct {
	BatchID   string    `json:"batch_id"`
	DataKey   string    `json:"data_key"`
	EventIDs  []string  `json:"event_ids"`
	Count     int       `json:"count"`
	CreatedAt time.Time `json:"created_at"`
}

func batchID(eventIDs []string, claimedAt time.Time, schemaVersion uint32) string {

	sorted := make([]string, len(eventIDs))

	copy(sorted, eventIDs)

	sort.Strings(sorted)

	h := sha256.New()
	for _, id := range sorted {
		h.Write([]byte(id))
	}
	h.Write([]byte(claimedAt.Truncate(time.Hour).Format(time.RFC3339)))

	_ = binary.Write(h, binary.BigEndian, schemaVersion) //ここはエラーにならない。lint回避。

	return hex.EncodeToString(h.Sum(nil))[:16] //返す
}

func s3DataKey(prefix string, claimedAt time.Time, bid string) string {
	return fmt.Sprintf("raw/%s/year=%04d/month=%02d/day=%02d/hour=%02d/%s.jsonl",
		prefix,
		claimedAt.Year(), claimedAt.Month(), claimedAt.Day(), claimedAt.Hour(),

		bid,
	)
}

func s3ManifestKey(prefix string, claimedAt time.Time, bid string) string {
	return fmt.Sprintf("raw/%s/year=%04d/month=%02d/day=%02d/hour=%02d/%s.manifest.json",
		prefix,
		claimedAt.Year(), claimedAt.Month(), claimedAt.Day(), claimedAt.Hour(),
		bid,
	)
}

func collectIDs(records []ClaimedEvent) []string {
	ids := make([]string, len(records))
	for i := range records {
		ids[i] = records[i].ID
	}
	return ids
}

func trimByByteLimit(records []ClaimedEvent, maxBytes int64) []ClaimedEvent {

	var total int64
	for i := range records {
		total += int64(len(records[i].Payload))
		if total > maxBytes {
			return records[:i]
		}
	}
	return records
}

func buildManifestJSON(m manifest) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")

	if err := enc.Encode(m); err != nil {
		return nil, fmt.Errorf("encode manifest: %w", err)
	}
	return buf.Bytes(), nil
}

func normalizePrefix(prefix string) string {
	return strings.TrimSuffix(strings.TrimPrefix(prefix, "/"), "/")
}
