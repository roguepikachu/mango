package generator

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	openai "github.com/sashabaranov/go-openai"

	"github.com/example/mango/internal/diff"
	"github.com/example/mango/internal/testmeta"
)

// Client abstracts the chat completion API.
type Client interface {
	ChatCompletion(ctx context.Context, prompt string) (string, error)
}

// OpenAIClient implements Client using OpenAI.
type OpenAIClient struct{ client *openai.Client }

// NewOpenAIClient returns a Client for OpenAI.
func NewOpenAIClient(token string) *OpenAIClient {
	if token == "" {
		return nil
	}
	return &OpenAIClient{client: openai.NewClient(token)}
}

// ChatCompletion sends a prompt and returns the assistant message.
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

// Generator creates new Ginkgo test scenarios via an LLM.
type Generator struct{ Client Client }

// New creates a Generator with the provided client.
func New(client Client) *Generator { return &Generator{Client: client} }

// Generate proposes new test scenario names.
func (g Generator) Generate(ctx context.Context, changes []diff.Change, tests []testmeta.Metadata) ([]string, error) {
	if g.Client == nil {
		return nil, fmt.Errorf("no client configured")
	}
	prompt := buildPrompt(changes, tests)
	out, err := g.Client.ChatCompletion(ctx, prompt)
	if err != nil {
		return nil, err
	}
	var names []string
	if err := json.Unmarshal([]byte(out), &names); err != nil {
		lines := strings.Split(out, "\n")
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

func buildPrompt(changes []diff.Change, tests []testmeta.Metadata) string {
	var b strings.Builder
	b.WriteString("Recent code changes:\n")
	for _, c := range changes {
		if len(c.Functions) > 0 {
			b.WriteString(fmt.Sprintf("- %s: %s\n", c.File, strings.Join(c.Functions, ", ")))
		} else {
			b.WriteString(fmt.Sprintf("- %s\n", c.File))
		}
	}
	b.WriteString("\nExisting tests:\n")
	for _, t := range tests {
		b.WriteString(fmt.Sprintf("- %s\n", t.Name))
	}
	b.WriteString("\nSuggest new Ginkgo test scenarios as a JSON array of names.")
	return b.String()
}
