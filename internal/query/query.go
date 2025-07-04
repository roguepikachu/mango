package query

import (
	"context"
	"fmt"
	"strings"

	openai "github.com/sashabaranov/go-openai"

	"github.com/example/mango/internal/testmeta"
)

// Client abstracts an LLM.
type Client interface {
	ChatCompletion(ctx context.Context, prompt string) (string, error)
}

// OpenAIClient implements Client with OpenAI.
type OpenAIClient struct{ client *openai.Client }

func NewOpenAIClient(token string) *OpenAIClient {
	if token == "" {
		return nil
	}
	return &OpenAIClient{client: openai.NewClient(token)}
}

func (c *OpenAIClient) ChatCompletion(ctx context.Context, prompt string) (string, error) {
	req := openai.ChatCompletionRequest{
		Model:     openai.GPT3Dot5Turbo,
		Messages:  []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: prompt}},
		MaxTokens: 512,
	}
	resp, err := c.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("empty response")
	}
	return resp.Choices[0].Message.Content, nil
}

// Service answers natural language queries about tests.
type Service struct{ Client Client }

func New(client Client) *Service { return &Service{Client: client} }

// Ask summarises tests matching a question.
func (s Service) Ask(ctx context.Context, question string, tests []testmeta.Metadata) (string, error) {
	if s.Client == nil {
		return "", fmt.Errorf("no client configured")
	}
	var b strings.Builder
	b.WriteString("Available tests:\n")
	for _, t := range tests {
		b.WriteString("- " + t.Name + "\n")
	}
	b.WriteString("\nQuestion: " + question)

	resp, err := s.Client.ChatCompletion(ctx, b.String())
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(resp), nil
}
