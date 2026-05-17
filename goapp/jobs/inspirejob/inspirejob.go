package inspirejob

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/app/job"
)

type InspireJob struct{}

const jobName = "inspire"

var quotes = []string{
	"When there is no desire, all things are at peace. - Laozi",
	"Simplicity is the ultimate sophistication. - Leonardo da Vinci",
	"Simplicity is the essence of happiness. - Cedric Bledsoe",
	"Smile, breathe, and go slowly. - Thich Nhat Hanh",
	"Simplicity is an acquired taste. - Katharine Gerould",
}

func NewInspireJob() *InspireJob {
	return &InspireJob{}
}

func (j *InspireJob) Name() string { return jobName }

func (j *InspireJob) Timeout() time.Duration { return 5 * time.Second }

func (j *InspireJob) Run(ctx context.Context, jobCtx job.JobContext, _ any) job.Result {
	quote := quotes[rand.Intn(len(quotes))]
	fmt.Printf("\n%s\n\n", quote)
	return job.Success(app.H{"quote": quote})
}
