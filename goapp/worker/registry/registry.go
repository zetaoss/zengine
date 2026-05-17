package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/zetaoss/zengine/goapp/app/appctx"
	"github.com/zetaoss/zengine/goapp/app/job"
	"github.com/zetaoss/zengine/goapp/jobs/commonreportjob"
	"github.com/zetaoss/zengine/goapp/jobs/editbotjob"
	"github.com/zetaoss/zengine/goapp/jobs/inspirejob"
	"github.com/zetaoss/zengine/goapp/jobs/pingdbjob"
	"github.com/zetaoss/zengine/goapp/jobs/pingredisjob"
	"github.com/zetaoss/zengine/goapp/jobs/runboxjob"
	"github.com/zetaoss/zengine/goapp/jobs/stat/cfjob"
	"github.com/zetaoss/zengine/goapp/jobs/stat/gajob"
	"github.com/zetaoss/zengine/goapp/jobs/stat/gscjob"
	"github.com/zetaoss/zengine/goapp/jobs/stat/mwjob"
	"github.com/zetaoss/zengine/goapp/jobs/writerequestjob"
	"github.com/zetaoss/zengine/goapp/worker/scheduler/schedule"
)

type Spec struct {
	Job        job.Job
	Queue      string
	Schedule   schedule.Schedule
	MaxRetries int
	Timeout    time.Duration
	Input      any
}

type option func(*Spec)

func withSchedule(s schedule.Schedule) option {
	return func(spec *Spec) {
		spec.Schedule = s
	}
}

func withQueue(q string) option {
	return func(spec *Spec) {
		spec.Queue = q
	}
}

type Registry struct {
	byName map[string]job.Job
	specs  map[string]Spec
}

func New() *Registry {
	r := &Registry{
		byName: make(map[string]job.Job),
		specs:  make(map[string]Spec),
	}

	register(r, cfjob.NewDailyJob(), withSchedule(schedule.HourlyAt(5)))
	register(r, cfjob.NewHourlyJob(), withSchedule(schedule.HourlyAt(5)))
	register(r, gajob.NewDailyJob(), withSchedule(schedule.HourlyAt(5)))
	register(r, gajob.NewHourlyJob(), withSchedule(schedule.HourlyAt(5)))
	register(r, gscjob.NewDailyJob(), withSchedule(schedule.HourlyAt(5)))
	register(r, gscjob.NewHourlyJob(), withSchedule(schedule.HourlyAt(5)))
	register(r, mwjob.NewDailyJob(), withSchedule(schedule.HourlyAt(5)))
	register(r, mwjob.NewHourlyJob(), withSchedule(schedule.HourlyAt(5)))
	register(r, commonreportjob.NewCommonReportJob())
	register(r, commonreportjob.NewNannyJob(), withSchedule(schedule.HourlyAt(0)))
	register(r, editbotjob.NewEditBotJob())
	register(r, editbotjob.NewNannyJob(), withSchedule(schedule.HourlyAt(0)))
	register(r, inspirejob.NewInspireJob())
	register(r, pingdbjob.NewPingDBJob())
	register(r, pingredisjob.NewPingRedisJob())
	register(r, runboxjob.NewPrunerJob())
	register(r, runboxjob.NewRunboxJob(), withQueue("runbox"))
	register(r, writerequestjob.NewPrunerJob(), withSchedule(schedule.DailyAt(0, 0)))
	register(r, writerequestjob.NewMatcherJob(), withSchedule(schedule.HourlyAt(15)))

	return r
}

func register[I any](r *Registry, tj job.TypedJob[I], opts ...option) {
	boundJob := job.Bind(tj)
	name := boundJob.Name()

	spec := Spec{Job: boundJob}
	for _, opt := range opts {
		opt(&spec)
	}

	r.byName[name] = boundJob
	r.specs[name] = spec
}

func (r *Registry) AllSpecs() []Spec {
	list := make([]Spec, 0, len(r.specs))
	for _, s := range r.specs {
		list = append(list, s)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Job.Name() < list[j].Job.Name()
	})
	return list
}

func (r *Registry) Find(name string) (job.Job, bool) {
	job, ok := r.byName[name]
	return job, ok
}

func (r *Registry) Run(ctx context.Context, appCtx *appctx.AppContext, req job.Request) ([]byte, error) {
	jobItem, ok := r.Find(req.JobName)
	if !ok {
		return nil, fmt.Errorf("unknown job: %s", req.JobName)
	}

	spec := r.specs[req.JobName]
	timeout := spec.Timeout
	if timeout == 0 {
		timeout = jobItem.Timeout()
	}
	if timeout == 0 {
		timeout = 1 * time.Hour // Global fallback
	}

	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	res := jobItem.Run(runCtx, appCtx, req.Input)

	b, _ := json.Marshal(res)
	if res.Status == job.StatusError {
		return b, fmt.Errorf("%v", res.Error)
	}
	return b, nil
}

func (r *Registry) Names() []string {
	names := make([]string, 0, len(r.byName))
	for name := range r.byName {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (r *Registry) Jobs() []job.Job {
	list := make([]job.Job, 0, len(r.byName))
	for _, job := range r.byName {
		list = append(list, job)
	}
	return list
}

func ConsumeQueues(specs []Spec) []string {
	seen := map[string]struct{}{}
	queues := make([]string, 0, 4)
	for _, item := range specs {
		queue := item.Queue
		if queue == "" {
			queue = "default"
		}
		if _, ok := seen[queue]; ok {
			continue
		}
		seen[queue] = struct{}{}
		queues = append(queues, queue)
	}
	sort.Strings(queues)
	return queues
}
