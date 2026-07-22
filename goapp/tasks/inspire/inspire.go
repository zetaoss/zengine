package inspire

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/zetaoss/zengine/goapp/app"
	"github.com/zetaoss/zengine/goapp/app/taskctx"
)

type InspireTask struct{}

var quotes = []string{
	"When there is no desire, all things are at peace. - Laozi",
	"Simplicity is the ultimate sophistication. - Leonardo da Vinci",
	"Simplicity is the essence of happiness. - Cedric Bledsoe",
	"Smile, breathe, and go slowly. - Thich Nhat Hanh",
	"Simplicity is an acquired taste. - Katharine Gerould",
}

func NewInspireTask() *InspireTask {
	return &InspireTask{}
}

func (j *InspireTask) Execute(ctx context.Context, taskCtx taskctx.Context, _ any) (app.H, error) {
	quote := quotes[rand.Intn(len(quotes))]
	fmt.Printf("\n%s\n\n", quote)
	return app.H{"quote": quote}, nil
}
