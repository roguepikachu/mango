# manGO

manGO is a smart test runner for Go projects. It analyses recent Git changes, extracts test metadata and consults an LLM to decide which tests should run.

## Usage

Build the CLI:

```bash
go build ./cmd/mango
```

Run the tool to automatically select and execute tests (diffing against `HEAD~1` by default):

```bash
./mango run
```

By default manGO uses OpenAI for test selection. Use `--provider` to choose `openai`, `anthropic` or `gemini`.

Preview tests selected without executing them:

```bash
./mango dry-run --diff HEAD~2
```

### CLI Flags

```
mango run [flags]

Flags:
  --diff string      Git diff range (default "HEAD~1")
  --mode string      Test backend: auto, go or ginkgo (default "auto")
  --llm-token string LLM API token (can also be set via LLM_TOKEN env var)
  --provider string   LLM provider: openai, anthropic, gemini (default "openai")
  --verbose          Enable debug logging
```

### Additional Commands

manGO offers extra functionality powered by LLMs:

```bash
# Generate missing Ginkgo scenarios from recent changes
mango generate-tests

# Predict which tests might fail based on an upcoming plan
mango predict --plan "describe feature work"

# Ask for code quality advice after running tests
mango advise

# Query tests using natural language
mango query --question "tests touching database layer"
```

### Makefile helpers

Common tasks are available via Makefile:

```bash
make build   # build the mango CLI
make test    # run unit tests
make e2e     # run end-to-end tests
```

End-to-end tests are guarded by the `e2e` build tag. Running `make e2e` or
`go test -tags=e2e ./...` will execute them.

## Project Layout

- `cmd/mango` - CLI entrypoint using Cobra
- `internal/diff` - git diff analysis
- `internal/testmeta` - test metadata extraction
- `internal/llmselector` - LLM based test selector
- `internal/executor` - test execution helpers
- `internal/orchestrator` - orchestrates the workflow
- `internal/generator` - intelligent scenario generation
- `internal/predictor` - predictive test execution
- `internal/advisor` - code quality advisor
- `internal/query` - natural language test querying
- `pkg/utils` - shared utilities
- `githooks` - example git hooks
- `k8s` - deployment configuration
- `Dockerfile` - container build

## Git Hooks

This project includes optional Git hooks in `githooks/`.
Install them with:

```bash
mkdir -p .git/hooks
cp githooks/* .git/hooks/
chmod +x .git/hooks/*
```
