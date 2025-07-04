package advisor

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

// Client abstracts chat completion.
type Client interface {
	ChatCompletion(ctx context.Context, prompt string) (string, error)
}

// OpenAIClient uses OpenAI for suggestions.
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

// Advisor provides code quality advice based on vet/test results.
type Advisor struct {
	Client Client
	Run    func(ctx context.Context, name string, args ...string) ([]byte, error)
}

func New(client Client) *Advisor {
	return &Advisor{Client: client, Run: func(ctx context.Context, name string, args ...string) ([]byte, error) {
		return exec.CommandContext(ctx, name, args...).CombinedOutput()
	}}
}

// Advise runs go vet and go test, then asks the LLM for suggestions.
func (a Advisor) Advise(ctx context.Context) (string, error) {
	if a.Client == nil {
		return "", fmt.Errorf("no client configured")
	}
	run := a.Run
	if run == nil {
		run = func(ctx context.Context, name string, args ...string) ([]byte, error) {
			return exec.CommandContext(ctx, name, args...).CombinedOutput()
		}
	}
	vetOut, _ := run(ctx, "go", "vet", "./...")
	testOut, _ := run(ctx, "go", "test", "./...")

	var b strings.Builder
	b.WriteString("go vet output:\n")
	b.Write(vetOut)
	b.WriteString("\n\nTest output:\n")
	b.Write(testOut)
	b.WriteString("\nProvide refactoring suggestions and code quality advice.")

	resp, err := a.Client.ChatCompletion(ctx, b.String())
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(resp), nil
}
