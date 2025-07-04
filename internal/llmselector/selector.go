package llmselector

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"

	"github.com/example/mango/internal/diff"
	"github.com/example/mango/internal/testmeta"
)

// Selector chooses relevant tests using an LLM.
//
//go:generate counterfeiter . Selector
type Selector interface {
	Select(ctx context.Context, changes []diff.Change, tests []testmeta.Metadata) ([]testmeta.Metadata, error)
}

// OpenAISelector implements Selector using the OpenAI API.
// Provider represents an LLM provider.
type Provider string

const (
	ProviderOpenAI    Provider = "openai"
	ProviderAnthropic Provider = "anthropic"
	ProviderGemini    Provider = "gemini"
)

// NewSelector returns a Selector for the given provider.
func NewSelector(provider Provider, token string) Selector {
	switch provider {
	case ProviderAnthropic:
		return NewAnthropicSelector(token)
	case ProviderGemini:
		return NewGeminiSelector(token)
	case ProviderOpenAI:
		fallthrough
	default:
		return NewOpenAISelector(token)
	}
}

// OpenAISelector implements Selector using the OpenAI API.
type OpenAISelector struct {
	Client *openai.Client
}

// NewOpenAISelector creates an OpenAI-based selector. If token is empty, nil is returned.
func NewOpenAISelector(token string) *OpenAISelector {
	if token == "" {
		return nil
	}
	c := openai.NewClient(token)
	return &OpenAISelector{Client: c}
}

// AnthropicSelector implements Selector using the Anthropic API.
type AnthropicSelector struct {
	Token  string
	Client *http.Client
	Model  string
}

// NewAnthropicSelector creates a selector for Anthropic Claude.
func NewAnthropicSelector(token string) *AnthropicSelector {
	if token == "" {
		return nil
	}
	return &AnthropicSelector{Token: token, Client: &http.Client{Timeout: 60 * time.Second}, Model: "claude-3-opus-20240229"}
}

// GeminiSelector implements Selector using the Gemini API.
type GeminiSelector struct {
	Token  string
	Client *http.Client
	Model  string
}

// NewGeminiSelector creates a selector for Google's Gemini.
func NewGeminiSelector(token string) *GeminiSelector {
	if token == "" {
		return nil
	}
	return &GeminiSelector{Token: token, Client: &http.Client{Timeout: 60 * time.Second}, Model: "gemini-pro"}
}

// Select asks the LLM which tests to run based on changes.
func (o *OpenAISelector) Select(ctx context.Context, changes []diff.Change, tests []testmeta.Metadata) ([]testmeta.Metadata, error) {
	if o == nil || o.Client == nil {
		return tests, nil
	}

	prompt := buildPrompt(changes, tests)
	req := openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{{
			Role:    openai.ChatMessageRoleUser,
			Content: prompt,
		}},
	}
	resp, err := o.Client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, err
	}
	if len(resp.Choices) == 0 {
		return nil, errors.New("no choices returned")
	}
	content := resp.Choices[0].Message.Content
	names, err := parseResponse(content)
	if err != nil {
		return nil, err
	}

	var selected []testmeta.Metadata
	for _, n := range names {
		for _, t := range tests {
			if strings.EqualFold(t.Name, n) {
				selected = append(selected, t)
			}
		}
	}
	if len(selected) == 0 {
		// fallback run all tests
		return tests, nil
	}
	return selected, nil
}

// Select asks Anthropic which tests to run.
func (a *AnthropicSelector) Select(ctx context.Context, changes []diff.Change, tests []testmeta.Metadata) ([]testmeta.Metadata, error) {
	if a == nil || a.Token == "" {
		return tests, nil
	}

	prompt := buildPrompt(changes, tests)
	body := map[string]interface{}{
		"model":      a.Model,
		"max_tokens": 512,
		"messages": []map[string]string{{
			"role":    "user",
			"content": prompt,
		}},
	}
	data, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.anthropic.com/v1/messages", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", a.Token)
	req.Header.Set("anthropic-version", "2023-06-01")
	resp, err := a.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var out struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if len(out.Content) == 0 {
		return nil, errors.New("empty response")
	}
	names, err := parseResponse(out.Content[0].Text)
	if err != nil {
		return nil, err
	}
	return filterTests(names, tests), nil
}

// Select asks Gemini which tests to run.
func (g *GeminiSelector) Select(ctx context.Context, changes []diff.Change, tests []testmeta.Metadata) ([]testmeta.Metadata, error) {
	if g == nil || g.Token == "" {
		return tests, nil
	}

	prompt := buildPrompt(changes, tests)
	body := map[string]interface{}{
		"contents": []map[string]interface{}{
			{"parts": []map[string]string{{"text": prompt}}},
		},
	}
	data, _ := json.Marshal(body)
	endpoint := "https://generativelanguage.googleapis.com/v1beta/models/" + url.PathEscape(g.Model) + ":generateContent?key=" + url.QueryEscape(g.Token)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := g.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var out struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if len(out.Candidates) == 0 || len(out.Candidates[0].Content.Parts) == 0 {
		return nil, errors.New("empty response")
	}
	names, err := parseResponse(out.Candidates[0].Content.Parts[0].Text)
	if err != nil {
		return nil, err
	}
	return filterTests(names, tests), nil
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
	b.WriteString("\nAvailable tests:\n")
	for i, t := range tests {
		b.WriteString(fmt.Sprintf("%d. %s\n", i+1, t.Name))
	}
	b.WriteString("\nRespond with a JSON array of test names to run.")
	return b.String()
}

func parseResponse(resp string) ([]string, error) {
	var names []string
	if err := json.Unmarshal([]byte(resp), &names); err == nil {
		return names, nil
	}

	// fallback: split by lines
	lines := strings.Split(resp, "\n")
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l != "" {
			names = append(names, l)
		}
	}
	if len(names) == 0 {
		return nil, errors.New("could not parse response")
	}
	return names, nil
}

func filterTests(names []string, all []testmeta.Metadata) []testmeta.Metadata {
	var selected []testmeta.Metadata
	for _, n := range names {
		for _, t := range all {
			if strings.EqualFold(t.Name, n) {
				selected = append(selected, t)
			}
		}
	}
	if len(selected) == 0 {
		return all
	}
	return selected
}
