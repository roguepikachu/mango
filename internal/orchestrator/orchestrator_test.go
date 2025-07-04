//go:build e2e
// +build e2e

package orchestrator

import (
	"context"
	"os"
	"os/exec"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/example/mango/internal/llmselector/llmselectorfakes"
	"github.com/example/mango/internal/testmeta"
)

var _ = Describe("Orchestrator", func() {
	It("runs dry-run workflow", func() {
		if _, err := exec.LookPath("git"); err != nil {
			Skip("git not installed")
		}
		dir := GinkgoT().TempDir()
		os.Chdir(dir)
		exec.Command("git", "init").Run()
		exec.Command("git", "config", "user.email", "a@b.c").Run()
		exec.Command("git", "config", "user.name", "t").Run()
		os.WriteFile("go.mod", []byte("module example.com/test\ngo 1.23.0"), 0o644)
		os.WriteFile("foo.go", []byte("package main\nfunc Add(a,b int) int { return a+b }"), 0o644)
		os.WriteFile("foo_test.go", []byte("package main\nimport \"testing\"\nfunc TestAdd(t *testing.T){}"), 0o644)
		exec.Command("git", "add", ".").Run()
		exec.Command("git", "commit", "-m", "init").Run()
		os.WriteFile("foo.go", []byte("package main\nfunc Add(a,b int) int { return a+b+1 }"), 0o644)
		exec.Command("git", "add", "foo.go").Run()
		exec.Command("git", "commit", "-m", "update").Run()

		meta, err := testmeta.Extract()
		Expect(err).NotTo(HaveOccurred())
		sel := &llmselectorfakes.FakeSelector{}
		sel.SelectReturns(meta, nil)
		orch := Orchestrator{Selector: sel, Mode: "auto", DryRun: true}
		err = orch.Run(context.Background(), "HEAD~1")
		Expect(err).NotTo(HaveOccurred())
	})
})

func TestOrchestrator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Orchestrator Suite")
}
