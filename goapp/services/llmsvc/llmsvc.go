package llmsvc

import (
	"context"
	"strings"

	"github.com/zetaoss/zengine/goapp/app/config"
	"github.com/zetaoss/zengine/goapp/services/llmsvc/client"
)

type Input struct {
	Prompt string
	Model  string
}

type Output struct {
	Content string
	Model   string
}

type LLMService struct {
	client *client.LLMClient
}

func New(cfg *config.Config) *LLMService {
	return &LLMService{client: client.New(cfg)}
}

func (s *LLMService) Generate(ctx context.Context, in Input) (Output, error) {
	content, model, err := s.client.ChatCompletion(ctx, in.Model, []client.Message{{Role: "user", Content: strings.TrimSpace(in.Prompt)}})
	if err != nil {
		return Output{}, err
	}
	return Output{Content: content, Model: model}, nil
}
