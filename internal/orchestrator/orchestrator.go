package orchestrator

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/example/mango/internal/diff"
	"github.com/example/mango/internal/executor"
	"github.com/example/mango/internal/llmselector"
	"github.com/example/mango/internal/testmeta"
)

// Orchestrator coordinates diff analysis, selection and execution.
type Orchestrator struct {
	Selector llmselector.Selector
	Mode     string // auto, go, ginkgo
	DryRun   bool
}

// Run performs the end-to-end workflow.
func (o Orchestrator) Run(ctx context.Context, diffRange string) error {
	changes, err := diff.AnalyzeDiff(diffRange)
	if err != nil {
		return err
	}

	tests, err := testmeta.Extract()
	if err != nil {
		return err
	}

	selected, err := o.Selector.Select(ctx, changes, tests)
	if err != nil {
		return err
	}

	fmt.Println("Selected tests:")
	for _, t := range selected {
		fmt.Printf("- %s (%s)\n", t.Name, t.File)
	}
	if o.DryRun {
		return nil
	}

	// group by package
	packages := map[string][]testmeta.Metadata{}
	for _, t := range selected {
		pkg := filepath.Dir(t.File)
		packages[pkg] = append(packages[pkg], t)
	}

	for pkg, metas := range packages {
		names := make([]string, len(metas))
		ginkgo := false
		for i, m := range metas {
			names[i] = m.Name
			if m.Ginkgo {
				ginkgo = true
			}
		}

		mode := o.Mode
		if mode == "auto" {
			if ginkgo {
				mode = "ginkgo"
			} else {
				mode = "go"
			}
		}

		switch mode {
		case "go":
			if err := executor.RunGoTests(ctx, pkg, names); err != nil {
				return err
			}
		case "ginkgo":
			if err := executor.RunGinkgo(ctx, pkg, names); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown mode %s", mode)
		}
	}

	return nil
}
