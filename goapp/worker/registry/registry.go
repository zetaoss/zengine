package registry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/hibiken/asynq"
	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/app/appctx"
	"github.com/zetaoss/zengine/goapp/app/taskctx"
	"github.com/zetaoss/zengine/goapp/tasks/aiedit"
	"github.com/zetaoss/zengine/goapp/tasks/commonreport"
	"github.com/zetaoss/zengine/goapp/tasks/inspire"
	"github.com/zetaoss/zengine/goapp/tasks/pingdb"
	"github.com/zetaoss/zengine/goapp/tasks/pingredis"
	"github.com/zetaoss/zengine/goapp/tasks/runbox"
	"github.com/zetaoss/zengine/goapp/tasks/stat/cf"
	"github.com/zetaoss/zengine/goapp/tasks/stat/ga"
	"github.com/zetaoss/zengine/goapp/tasks/stat/gsc"
	"github.com/zetaoss/zengine/goapp/tasks/stat/mw"
	"github.com/zetaoss/zengine/goapp/tasks/writerequest"
)

type processor func(context.Context, taskctx.Context, []byte) (app.H, error)

var errInvalidPayload = errors.New("invalid task payload")

type Spec struct {
	Type       string
	Queue      string
	Cron       string
	MaxRetries int
	Timeout    time.Duration
	Input      any
	process    processor
}

type executor[I any] interface {
	Execute(context.Context, taskctx.Context, I) (app.H, error)
}

type Registry struct{ specs map[string]Spec }

func New() *Registry {
	r := &Registry{specs: make(map[string]Spec)}
	register(r, "stat-cf-daily", 5*time.Minute, cf.NewDailyTask(), cron("5 * * * *"))
	register(r, "stat-cf-hourly", 5*time.Minute, cf.NewHourlyTask(), cron("5 * * * *"))
	register(r, "stat-ga-daily", 5*time.Minute, ga.NewDailyTask(), cron("5 * * * *"))
	register(r, "stat-ga-hourly", 5*time.Minute, ga.NewHourlyTask(), cron("5 * * * *"))
	register(r, "stat-gsc-daily", 5*time.Minute, gsc.NewDailyTask(), cron("5 * * * *"))
	register(r, "stat-gsc-hourly", 5*time.Minute, gsc.NewHourlyTask(), cron("5 * * * *"))
	register(r, "stat-mw-daily", 5*time.Minute, mw.NewDailyTask(), cron("5 * * * *"))
	register(r, "stat-mw-hourly", 5*time.Minute, mw.NewHourlyTask(), cron("5 * * * *"))
	register(r, "common-report", 5*time.Minute, commonreport.NewCommonReportTask())
	register(r, "common-report-nanny", time.Minute, commonreport.NewNannyTask(), cron("0 * * * *"))
	register(r, "ai-edit", 10*time.Minute, aiedit.NewAIEditTask())
	register(r, "ai-edit-nanny", time.Minute, aiedit.NewNannyTask(), cron("0 * * * *"))
	register(r, "inspire", 5*time.Second, inspire.NewInspireTask())
	register(r, "ping-db", 10*time.Second, pingdb.NewPingDBTask())
	register(r, "ping-redis", 5*time.Second, pingredis.NewPingRedisTask())
	register(r, "runbox-pruner", time.Minute, runbox.NewPrunerTask())
	register(r, "runbox", 5*time.Minute, runbox.NewRunboxTask(), queue("runbox"))
	register(r, "request-pruner", 5*time.Minute, writerequest.NewPrunerTask(), cron("0 0 * * *"))
	register(r, "request-matcher", 5*time.Minute, writerequest.NewMatcherTask(), cron("15 * * * *"))
	return r
}

type option func(*Spec)

func cron(value string) option  { return func(s *Spec) { s.Cron = value } }
func queue(value string) option { return func(s *Spec) { s.Queue = value } }

func register[I any](r *Registry, taskType string, timeout time.Duration, task executor[I], opts ...option) {
	spec := Spec{Type: taskType, Timeout: timeout, MaxRetries: 3}
	for _, opt := range opts {
		opt(&spec)
	}
	spec.process = func(ctx context.Context, tc taskctx.Context, raw []byte) (app.H, error) {
		var payload I
		if len(raw) > 0 {
			if err := json.Unmarshal(raw, &payload); err != nil {
				return nil, fmt.Errorf("%w: decode %s: %v", errInvalidPayload, taskType, err)
			}
		}
		return task.Execute(ctx, tc, payload)
	}
	r.specs[taskType] = spec
}

func (r *Registry) AllSpecs() []Spec {
	list := make([]Spec, 0, len(r.specs))
	for _, spec := range r.specs {
		list = append(list, spec)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Type < list[j].Type })
	return list
}
func (r *Registry) FindSpec(taskType string) (Spec, bool) {
	spec, ok := r.specs[taskType]
	return spec, ok
}
func (r *Registry) Names() []string {
	specs := r.AllSpecs()
	names := make([]string, len(specs))
	for i := range specs {
		names[i] = specs[i].Type
	}
	return names
}

func (r *Registry) Process(ctx context.Context, appCtx *appctx.AppContext, task *asynq.Task) error {
	spec, ok := r.FindSpec(task.Type())
	if !ok {
		return fmt.Errorf("unknown task: %s", task.Type())
	}
	_, err := spec.process(ctx, appCtx, task.Payload())
	if errors.Is(err, errInvalidPayload) {
		return fmt.Errorf("%w: %w", asynq.SkipRetry, err)
	}
	return err
}

func (r *Registry) RunDirect(ctx context.Context, appCtx *appctx.AppContext, taskType string, raw []byte) ([]byte, error) {
	spec, ok := r.FindSpec(taskType)
	if !ok {
		return nil, fmt.Errorf("unknown task: %s", taskType)
	}
	runCtx, cancel := context.WithTimeout(ctx, spec.Timeout)
	defer cancel()
	data, err := spec.process(runCtx, appCtx, raw)
	result := struct {
		Status string `json:"status"`
		Data   app.H  `json:"data,omitempty"`
		Error  string `json:"error,omitempty"`
	}{Status: "Success", Data: data}
	if err != nil {
		result.Status, result.Error = "Error", err.Error()
	}
	b, marshalErr := json.Marshal(result)
	if marshalErr != nil {
		return nil, marshalErr
	}
	return b, err
}

func NewTask(spec Spec, input any) (*asynq.Task, []asynq.Option, error) {
	raw, err := json.Marshal(input)
	if err != nil {
		return nil, nil, fmt.Errorf("encode %s payload: %w", spec.Type, err)
	}
	q := spec.Queue
	if q == "" {
		q = "default"
	}
	return asynq.NewTask(spec.Type, raw), []asynq.Option{asynq.Queue(q), asynq.MaxRetry(spec.MaxRetries), asynq.Timeout(spec.Timeout)}, nil
}

func ConsumeQueues(specs []Spec) []string {
	seen := map[string]struct{}{}
	queues := make([]string, 0, 4)
	for _, spec := range specs {
		q := spec.Queue
		if q == "" {
			q = "default"
		}
		if _, ok := seen[q]; !ok {
			seen[q] = struct{}{}
			queues = append(queues, q)
		}
	}
	sort.Strings(queues)
	return queues
}
