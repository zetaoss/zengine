package job

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/app/config"

	"gorm.io/gorm"
)

// Context is the interface that provides necessary environment for a job.
// This breaks the circular dependency between app.AppContext and the job package.
type JobContext interface {
	GetDB() (*gorm.DB, error)
	Config() *config.Config
	Enqueue(ctx context.Context, req Request) (uint64, error)
}

// Job is the non-generic interface used by the Registry and Worker.
type Job interface {
	Name() string
	Timeout() time.Duration
	Decode(raw []byte) (any, error)
	Run(ctx context.Context, jobCtx JobContext, input any) Result
}

// TypedJob is the generic interface that individual jobs should implement.
type TypedJob[I any] interface {
	Name() string
	Timeout() time.Duration
	Run(ctx context.Context, jobCtx JobContext, input I) Result
}

// Bind wraps a TypedJob into a non-generic Job interface for the Registry.
func Bind[I any](tj TypedJob[I]) Job {
	return &jobAdapter[I]{inner: tj}
}

type jobAdapter[I any] struct {
	inner TypedJob[I]
}

func (a *jobAdapter[I]) Name() string { return a.inner.Name() }

func (a *jobAdapter[I]) Timeout() time.Duration { return a.inner.Timeout() }

func (a *jobAdapter[I]) Decode(raw []byte) (any, error) {
	var input I
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &input); err != nil {
			return nil, fmt.Errorf("decode %s: %w", a.Name(), err)
		}
	}
	return input, nil
}

func (a *jobAdapter[I]) Run(ctx context.Context, jobCtx JobContext, input any) Result {
	if input == nil {
		var zero I
		return a.inner.Run(ctx, jobCtx, zero)
	}
	typedInput, ok := input.(I)
	if !ok {
		return Error(fmt.Errorf("invalid input type for job %s: expected %T, got %T", a.Name(), typedInput, input))
	}
	return a.inner.Run(ctx, jobCtx, typedInput)
}

type Request struct {
	ID          uint64
	JobName     string
	Queue       string
	Input       any
	Attempt     int
	MaxRetries  int
	RequestedAt time.Time
	RunAt       *time.Time
	DedupeKey   string
}

type Enqueuer interface {
	Enqueue(context.Context, Request) (uint64, error)
}

// Status represents the state of a job execution.
type Status string

const (
	StatusSuccess Status = "Success"
	StatusError   Status = "Error"
)

// ResultError is a wrapper for error that supports JSON marshaling.
type ResultError struct {
	err error
}

func (e *ResultError) Error() string {
	if e == nil || e.err == nil {
		return ""
	}
	return e.err.Error()
}

func (e *ResultError) MarshalJSON() ([]byte, error) {
	if e == nil || e.err == nil {
		return []byte("null"), nil
	}
	return json.Marshal(e.err.Error())
}

func (e *ResultError) Unwrap() error {
	return e.err
}

// Result is the standardized output format for all jobs.
type Result struct {
	Status Status       `json:"status"`
	Data   app.H        `json:"data,omitempty"`
	Error  *ResultError `json:"error,omitempty"`
}

// Success creates a standardized success result for a job.
func Success(data app.H) Result {
	return Result{
		Status: StatusSuccess,
		Data:   data,
	}
}

// Error creates a standardized error result for a job.
func Error(err error) Result {
	if err == nil {
		return Result{Status: StatusError}
	}
	return Result{
		Status: StatusError,
		Error:  &ResultError{err: err},
	}
}
