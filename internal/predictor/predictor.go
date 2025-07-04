package predictor

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	openai "github.com/sashabaranov/go-openai"

	"github.com/example/mango/internal/testmeta"
)

// Client abstracts the LLM.
type Client interface {
	ChatCompletion(ctx context.Context, prompt string) (string, error)
}

// OpenAIClient implements Client using OpenAI.
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

// Predictor forecasts future failing tests from planned changes.
type Predictor struct{ Client Client }

func New(client Client) *Predictor { return &Predictor{Client: client} }

// Predict returns tests likely to fail given an upcoming change description.
func (p Predictor) Predict(ctx context.Context, description string, tests []testmeta.Metadata) ([]string, error) {
	if p.Client == nil {
		return nil, fmt.Errorf("no client configured")
	}
	var b strings.Builder
	b.WriteString("Planned change:\n")
	b.WriteString(description)
	b.WriteString("\nAvailable tests:\n")
	for _, t := range tests {
		b.WriteString("- " + t.Name + "\n")
	}
	b.WriteString("\nWhich tests are most likely to fail? Respond with a JSON array of test names.")
	resp, err := p.Client.ChatCompletion(ctx, b.String())
	if err != nil {
		return nil, err
	}
	var names []string
	if err := json.Unmarshal([]byte(resp), &names); err != nil {
		lines := strings.Split(resp, "\n")
		for _, l := range lines {
			l = strings.TrimSpace(l)
			if l != "" {
				names = append(names, l)
			}
		}
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("could not parse response")
	}
	return names, nil
}
