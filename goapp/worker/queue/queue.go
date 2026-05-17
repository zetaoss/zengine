package queue

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/zetaoss/zengine/goapp/app/job"

	goredis "github.com/redis/go-redis/v9"
)

type RequestRow struct {
	ID          uint64
	JobName     string
	Queue       string
	Payload     []byte
	Status      string
	Attempt     int
	MaxRetries  int
	DedupeKey   string
	LastError   string
	RequestedAt time.Time
	RunAt       *time.Time
	LockedAt    *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type RunningJob struct {
	ID       uint64
	JobName  string
	Queue    string
	Attempt  int
	LockedAt *time.Time
	Payload  []byte
}

type PendingJob struct {
	ID          uint64
	JobName     string
	Queue       string
	Attempt     int
	RequestedAt time.Time
	RunAt       *time.Time
	Payload     []byte
}

type RedisQueue struct {
	rdb    *goredis.Client
	prefix string
}

func New(client *goredis.Client) *RedisQueue {
	return &RedisQueue{rdb: client, prefix: "zengine:jobq"}
}

func (q *RedisQueue) Enqueue(ctx context.Context, req job.Request) (uint64, error) {
	now := time.Now()
	queue := req.Queue
	if queue == "" {
		queue = "default"
	}
	maxRetries := req.MaxRetries
	if maxRetries < 0 {
		maxRetries = 0
	}
	runAt := now
	if req.RunAt != nil {
		runAt = req.RunAt.UTC()
	}

	if req.DedupeKey != "" {
		dedupe := q.keyDedupe(req.JobName, req.DedupeKey)
		existing, err := q.rdb.Get(ctx, dedupe).Result()
		if err == nil {
			id, convErr := strconv.ParseUint(existing, 10, 64)
			if convErr == nil {
				return id, nil
			}
		}
		if err != nil && err != goredis.Nil {
			return 0, err
		}
	}

	id, err := q.rdb.Incr(ctx, q.keyNextID()).Uint64()
	if err != nil {
		return 0, err
	}

	row := RequestRow{
		ID:          id,
		JobName:     req.JobName,
		Queue:       queue,
		Payload:     mustJSON(req.Input),
		Status:      "pending",
		Attempt:     0,
		MaxRetries:  maxRetries,
		DedupeKey:   req.DedupeKey,
		RequestedAt: now,
		RunAt:       new(runAt),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	pipe := q.rdb.TxPipeline()
	pipe.HSet(ctx, q.keyReq(id), encodeRow(row))
	pipe.Expire(ctx, q.keyReq(id), 24*time.Hour)
	pipe.ZAdd(ctx, q.keyPending(queue), goredis.Z{Score: float64(runAt.UnixMilli()), Member: id})
	if req.DedupeKey != "" {
		pipe.SetNX(ctx, q.keyDedupe(req.JobName, req.DedupeKey), strconv.FormatUint(id, 10), 0)
	}
	_, err = pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (q *RedisQueue) ClaimPending(queue string) (*RequestRow, error) {
	ctx := context.Background()
	if queue == "" {
		queue = "default"
	}

	nowMs := time.Now().UnixMilli()
	claimScript := goredis.NewScript(`
local pending = KEYS[1]
local now = ARGV[1]
local ids = redis.call('ZRANGEBYSCORE', pending, '-inf', now, 'LIMIT', 0, 1)
if #ids == 0 then
  return ''
end
local id = ids[1]
if redis.call('ZREM', pending, id) == 1 then
  return id
end
return ''
`)

	idStr, err := claimScript.Run(ctx, q.rdb, []string{q.keyPending(queue)}, nowMs).Text()
	if err == goredis.Nil || idStr == "" {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return nil, err
	}

	vals, err := q.rdb.HGetAll(ctx, q.keyReq(id)).Result()
	if err != nil {
		return nil, err
	}
	if len(vals) == 0 {
		return nil, nil
	}
	row := decodeRow(vals)
	now := time.Now()
	row.Status = "running"
	row.Attempt++
	row.LockedAt = new(now)
	row.UpdatedAt = now

	pipe := q.rdb.TxPipeline()
	pipe.HSet(ctx, q.keyReq(id), encodeRow(row))
	pipe.Expire(ctx, q.keyReq(id), 24*time.Hour)
	pipe.ZAdd(ctx, q.keyRunning(), goredis.Z{Score: float64(now.UnixMilli()), Member: id})
	_, err = pipe.Exec(ctx)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (q *RedisQueue) MarkSucceeded(id uint64) error {
	ctx := context.Background()
	vals, err := q.rdb.HGetAll(ctx, q.keyReq(id)).Result()
	if err != nil {
		return err
	}
	if len(vals) == 0 {
		return nil
	}
	row := decodeRow(vals)
	now := time.Now()
	row.Status = "succeeded"
	row.LockedAt = nil
	row.UpdatedAt = now

	pipe := q.rdb.TxPipeline()
	pipe.HSet(ctx, q.keyReq(id), encodeRow(row))
	pipe.ZRem(ctx, q.keyRunning(), id)
	_, err = pipe.Exec(ctx)
	return err
}

func (q *RedisQueue) MarkFailed(id uint64, runErr error) error {
	ctx := context.Background()
	vals, err := q.rdb.HGetAll(ctx, q.keyReq(id)).Result()
	if err != nil {
		return err
	}
	if len(vals) == 0 {
		return nil
	}
	row := decodeRow(vals)
	now := time.Now()
	row.Status = "failed"
	if runErr != nil {
		row.LastError = runErr.Error()
	}
	row.LockedAt = nil
	row.UpdatedAt = now

	pipe := q.rdb.TxPipeline()
	pipe.HSet(ctx, q.keyReq(id), encodeRow(row))
	pipe.ZRem(ctx, q.keyRunning(), id)
	_, err = pipe.Exec(ctx)
	return err
}

func (q *RedisQueue) MarkRetryOrFailed(row *RequestRow, runErr error) error {
	ctx := context.Background()
	now := time.Now()
	if runErr != nil {
		row.LastError = runErr.Error()
	}

	pipe := q.rdb.TxPipeline()
	if row.Attempt > row.MaxRetries {
		row.Status = "failed"
		row.LockedAt = nil
		row.UpdatedAt = now
		pipe.HSet(ctx, q.keyReq(row.ID), encodeRow(*row))
		pipe.ZRem(ctx, q.keyRunning(), row.ID)
	} else {
		next := now.Add(time.Duration(row.Attempt) * time.Second)
		row.Status = "pending"
		row.RunAt = &next
		row.LockedAt = nil
		row.UpdatedAt = now
		pipe.HSet(ctx, q.keyReq(row.ID), encodeRow(*row))
		pipe.ZRem(ctx, q.keyRunning(), row.ID)
		pipe.ZAdd(ctx, q.keyPending(row.Queue), goredis.Z{Score: float64(next.UnixMilli()), Member: row.ID})
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (q *RedisQueue) StartDirectRun(jobName string) (uint64, error) {
	ctx := context.Background()
	now := time.Now()
	id, err := q.rdb.Incr(ctx, q.keyNextID()).Uint64()
	if err != nil {
		return 0, err
	}
	row := RequestRow{
		ID:          id,
		JobName:     jobName,
		Queue:       "direct",
		Status:      "running",
		Attempt:     1,
		MaxRetries:  0,
		RequestedAt: now,
		RunAt:       new(now),
		LockedAt:    new(now),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	pipe := q.rdb.TxPipeline()
	pipe.HSet(ctx, q.keyReq(id), encodeRow(row))
	pipe.Expire(ctx, q.keyReq(id), 24*time.Hour)
	pipe.ZAdd(ctx, q.keyRunning(), goredis.Z{Score: float64(now.UnixMilli()), Member: id})
	_, err = pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (q *RedisQueue) ListRunning(limit int) ([]RunningJob, error) {
	ctx := context.Background()
	if limit <= 0 {
		limit = 100
	}
	ids, err := q.rdb.ZRangeArgs(ctx, goredis.ZRangeArgs{
		Key:   q.keyRunning(),
		Start: 0,
		Stop:  int64(limit - 1),
		Rev:   true,
	}).Result()
	if err != nil {
		return nil, err
	}
	out := make([]RunningJob, 0, len(ids))
	for _, idStr := range ids {
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			continue
		}
		vals, err := q.rdb.HGetAll(ctx, q.keyReq(id)).Result()
		if err != nil || len(vals) == 0 {
			continue
		}
		row := decodeRow(vals)
		if row.Status != "running" {
			continue
		}
		out = append(out, RunningJob{
			ID:       row.ID,
			JobName:  row.JobName,
			Queue:    row.Queue,
			Attempt:  row.Attempt,
			LockedAt: row.LockedAt,
			Payload:  row.Payload,
		})
	}
	return out, nil
}

func (q *RedisQueue) ListPending(queue string, limit int) ([]PendingJob, error) {
	ctx := context.Background()
	if queue == "" {
		queue = "default"
	}
	if limit <= 0 {
		limit = 100
	}
	ids, err := q.rdb.ZRangeArgs(ctx, goredis.ZRangeArgs{
		Key:   q.keyPending(queue),
		Start: 0,
		Stop:  int64(limit - 1),
	}).Result()
	if err != nil {
		return nil, err
	}
	out := make([]PendingJob, 0, len(ids))
	for _, idStr := range ids {
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			continue
		}
		vals, err := q.rdb.HGetAll(ctx, q.keyReq(id)).Result()
		if err != nil || len(vals) == 0 {
			continue
		}
		row := decodeRow(vals)
		if row.Status != "pending" {
			continue
		}
		out = append(out, PendingJob{
			ID:          row.ID,
			JobName:     row.JobName,
			Queue:       row.Queue,
			Attempt:     row.Attempt,
			RequestedAt: row.RequestedAt,
			RunAt:       row.RunAt,
			Payload:     row.Payload,
		})
	}
	return out, nil
}

func (q *RedisQueue) FlushRunning(reason string) (int64, error) {
	ctx := context.Background()
	msg := strings.TrimSpace(reason)
	if msg == "" {
		msg = "flushed by worker command"
	}
	ids, err := q.rdb.ZRangeArgs(ctx, goredis.ZRangeArgs{
		Key:   q.keyRunning(),
		Start: 0,
		Stop:  -1,
	}).Result()
	if err != nil {
		return 0, err
	}
	now := time.Now()
	pipe := q.rdb.TxPipeline()
	var n int64
	for _, idStr := range ids {
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			continue
		}
		vals, err := q.rdb.HGetAll(ctx, q.keyReq(id)).Result()
		if err != nil || len(vals) == 0 {
			continue
		}
		row := decodeRow(vals)
		if row.Status != "running" {
			continue
		}
		row.Status = "failed"
		row.LastError = fmt.Sprintf("%s at %s", msg, now.UTC().Format(time.RFC3339))
		row.LockedAt = nil
		row.UpdatedAt = now
		pipe.HSet(ctx, q.keyReq(id), encodeRow(row))
		n++
	}
	pipe.Del(ctx, q.keyRunning())
	if _, err := pipe.Exec(ctx); err != nil {
		return 0, err
	}
	return n, nil
}

func (q *RedisQueue) FlushPending(reason string) (int64, error) {
	ctx := context.Background()
	msg := strings.TrimSpace(reason)
	if msg == "" {
		msg = "flushed pending by worker command"
	}
	now := time.Now()
	var n int64
	pipe := q.rdb.TxPipeline()
	for _, queue := range []string{"default", "runbox"} {
		ids, err := q.rdb.ZRangeArgs(ctx, goredis.ZRangeArgs{
			Key:   q.keyPending(queue),
			Start: 0,
			Stop:  -1,
		}).Result()
		if err != nil {
			return n, err
		}
		for _, idStr := range ids {
			id, err := strconv.ParseUint(idStr, 10, 64)
			if err != nil {
				continue
			}
			vals, err := q.rdb.HGetAll(ctx, q.keyReq(id)).Result()
			if err != nil || len(vals) == 0 {
				continue
			}
			row := decodeRow(vals)
			if row.Status != "pending" {
				continue
			}
			row.Status = "failed"
			row.LastError = fmt.Sprintf("%s at %s", msg, now.UTC().Format(time.RFC3339))
			row.UpdatedAt = now
			pipe.HSet(ctx, q.keyReq(id), encodeRow(row))
			pipe.ZRem(ctx, q.keyPending(queue), row.ID)
			n++
		}
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return n, err
	}
	return n, nil
}

func (q *RedisQueue) keyNextID() string       { return q.prefix + ":next_id" }
func (q *RedisQueue) keyReq(id uint64) string { return fmt.Sprintf("%s:req:%d", q.prefix, id) }
func (q *RedisQueue) keyPending(queue string) string {
	return fmt.Sprintf("%s:pending:%s", q.prefix, queue)
}
func (q *RedisQueue) keyRunning() string { return q.prefix + ":running" }
func (q *RedisQueue) keyDedupe(job, dedupe string) string {
	return fmt.Sprintf("%s:dedupe:%s:%s", q.prefix, job, dedupe)
}

func encodeRow(row RequestRow) map[string]any {
	out := map[string]any{
		"id":           strconv.FormatUint(row.ID, 10),
		"job_name":     row.JobName,
		"queue":        row.Queue,
		"status":       row.Status,
		"attempt":      strconv.Itoa(row.Attempt),
		"max_retries":  strconv.Itoa(row.MaxRetries),
		"dedupe_key":   row.DedupeKey,
		"last_error":   row.LastError,
		"requested_at": strconv.FormatInt(row.RequestedAt.UnixMilli(), 10),
		"created_at":   strconv.FormatInt(row.CreatedAt.UnixMilli(), 10),
		"updated_at":   strconv.FormatInt(row.UpdatedAt.UnixMilli(), 10),
		"payload":      base64.StdEncoding.EncodeToString(row.Payload),
	}
	if row.RunAt != nil {
		out["run_at"] = strconv.FormatInt(row.RunAt.UnixMilli(), 10)
	} else {
		out["run_at"] = ""
	}
	if row.LockedAt != nil {
		out["locked_at"] = strconv.FormatInt(row.LockedAt.UnixMilli(), 10)
	} else {
		out["locked_at"] = ""
	}
	return out
}

func decodeRow(m map[string]string) RequestRow {
	id, _ := strconv.ParseUint(m["id"], 10, 64)
	attempt, _ := strconv.Atoi(m["attempt"])
	maxRetries, _ := strconv.Atoi(m["max_retries"])
	payload, _ := base64.StdEncoding.DecodeString(m["payload"])
	return RequestRow{
		ID:          id,
		JobName:     m["job_name"],
		Queue:       m["queue"],
		Payload:     payload,
		Status:      m["status"],
		Attempt:     attempt,
		MaxRetries:  maxRetries,
		DedupeKey:   m["dedupe_key"],
		LastError:   m["last_error"],
		RequestedAt: parseMillis(m["requested_at"]),
		RunAt:       parseMillisPtr(m["run_at"]),
		LockedAt:    parseMillisPtr(m["locked_at"]),
		CreatedAt:   parseMillis(m["created_at"]),
		UpdatedAt:   parseMillis(m["updated_at"]),
	}
}

func parseMillis(s string) time.Time {
	ms, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return time.Time{}
	}
	return time.UnixMilli(ms)
}

func parseMillisPtr(s string) *time.Time {
	if s == "" {
		return nil
	}
	t := parseMillis(s)
	if t.IsZero() {
		return nil
	}
	return &t
}

func new(t time.Time) *time.Time {
	return &t
}

func mustJSON(v any) []byte {
	if v == nil {
		return nil
	}
	b, _ := json.Marshal(v)
	return b
}
