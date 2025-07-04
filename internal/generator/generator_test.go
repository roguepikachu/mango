package generator

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/example/mango/internal/diff"
	"github.com/example/mango/internal/testmeta"
)

type fakeClient struct {
	resp string
	err  error
}

func (f fakeClient) ChatCompletion(ctx context.Context, prompt string) (string, error) {
	return f.resp, f.err
}

var _ = Describe("Generator", func() {
	It("parses LLM suggestions", func() {
		g := New(fakeClient{resp: `["A","B"]`})
		names, err := g.Generate(context.Background(), []diff.Change{{File: "a.go"}}, []testmeta.Metadata{})
		Expect(err).NotTo(HaveOccurred())
		Expect(names).To(ConsistOf("A", "B"))
	})
})

func TestGenerator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Generator Suite")
}
