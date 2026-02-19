package compaction

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/memory"
	"github.com/apache/arrow-go/v18/parquet/pqarrow"
)

var taskEventSchema = arrow.NewSchema([]arrow.Field{
	{Name: "id", Type: arrow.BinaryTypes.String, Nullable: false},
	{Name: "user_id", Type: arrow.BinaryTypes.String, Nullable: false},
	{Name: "task_id", Type: arrow.BinaryTypes.String, Nullable: false},
	{Name: "request_id", Type: arrow.BinaryTypes.String, Nullable: false},
	{Name: "event_type", Type: arrow.BinaryTypes.String, Nullable: false},
	{Name: "occurred_at", Type: arrow.FixedWidthTypes.Timestamp_us, Nullable: false},
	{Name: "schema_version", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
	{Name: "payload", Type: arrow.BinaryTypes.String, Nullable: false},
}, nil)

// writeParquet は events を Parquet バイト列に変換する。
func writeParquet(events []Event) ([]byte, error) {
	alloc := memory.NewGoAllocator()
	bldr := array.NewRecordBuilder(alloc, taskEventSchema)
	defer bldr.Release()

	idBldr := bldr.Field(0).(*array.StringBuilder)
	userIDBldr := bldr.Field(1).(*array.StringBuilder)
	taskIDBldr := bldr.Field(2).(*array.StringBuilder)
	requestIDBldr := bldr.Field(3).(*array.StringBuilder)
	eventTypeBldr := bldr.Field(4).(*array.StringBuilder)
	occurredAtBldr := bldr.Field(5).(*array.TimestampBuilder)
	schemaVerBldr := bldr.Field(6).(*array.Int32Builder)
	payloadBldr := bldr.Field(7).(*array.StringBuilder)

	for i := range events {
		idBldr.Append(events[i].ID)
		userIDBldr.Append(events[i].UserID)
		taskIDBldr.Append(events[i].TaskID)
		requestIDBldr.Append(events[i].RequestID)
		eventTypeBldr.Append(events[i].EventType)
		occurredAtBldr.Append(arrow.Timestamp(events[i].OccurredAt.UnixMicro()))
		schemaVerBldr.Append(int32(events[i].SchemaVersion))
		payloadBldr.Append(string(events[i].Payload))
	}

	rec := bldr.NewRecord()
	defer rec.Release()

	var buf bytes.Buffer
	writer, err := pqarrow.NewFileWriter(taskEventSchema, &buf, nil, pqarrow.DefaultWriterProps())
	if err != nil {
		return nil, fmt.Errorf("create parquet writer: %w", err)
	}
	if err := writer.Write(rec); err != nil {
		return nil, fmt.Errorf("write parquet record: %w", err)
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close parquet writer: %w", err)
	}
	return buf.Bytes(), nil
}

// compactedKey は compacted パーティションの S3 キーを生成する。
// claim 日ごとに別ファイル → 各 claim 日の compaction が独立して冪等。
func compactedKey(compactedPrefix string, occurredDay string, claimDate string) string {
	return fmt.Sprintf("%s/day=%s/from-claims-%s.parquet",
		compactedPrefix, occurredDay, claimDate,
	)
}

// uploadParquet は events を Parquet に変換して S3 にアップロードする。
func uploadParquet(
	ctx context.Context,
	storage ObjectStorage,
	compactedPrefix string,
	occurredDay string,
	claimDate string,
	events []Event,
	metrics *Metrics,
) error {
	writeStart := time.Now()
	data, err := writeParquet(events)
	metrics.ParquetWriteDuration.Add(time.Since(writeStart).Seconds())
	if err != nil {
		return fmt.Errorf("write parquet for day=%s: %w", occurredDay, err)
	}
	key := compactedKey(compactedPrefix, occurredDay, claimDate)
	uploadStart := time.Now()
	if err := storage.Upload(ctx, key, bytes.NewReader(data)); err != nil {
		return fmt.Errorf("upload parquet key=%s: %w", key, err)
	}
	metrics.S3UploadDuration.Add(time.Since(uploadStart).Seconds())
	return nil
}
