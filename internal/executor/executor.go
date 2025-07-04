package executor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// RunGoTests runs go tests matching the given regex in the specified package.
func RunGoTests(ctx context.Context, pkg string, tests []string) error {
	if len(tests) == 0 {
		return nil
	}
	regex := fmt.Sprintf("^(%s)$", strings.Join(tests, "|"))
	args := []string{"test", pkg, "-run", regex}
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// RunGinkgo runs ginkgo tests focusing on the provided expressions.
func RunGinkgo(ctx context.Context, pkg string, focuses []string) error {
	if len(focuses) == 0 {
		return nil
	}
	focus := strings.Join(focuses, "|")
	args := []string{"test", pkg, "-ginkgo.focus", focus}
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
