package predictor

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/example/mango/internal/testmeta"
)

type fakeClient struct {
	resp string
	err  error
}

func (f fakeClient) ChatCompletion(ctx context.Context, prompt string) (string, error) {
	return f.resp, f.err
}

var _ = Describe("Predictor", func() {
	It("parses predicted tests", func() {
		p := New(fakeClient{resp: `["A"]`})
		names, err := p.Predict(context.Background(), "future", []testmeta.Metadata{{Name: "A"}, {Name: "B"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(names).To(ConsistOf("A"))
	})
})

func TestPredictor(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Predictor Suite")
}
