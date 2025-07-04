package advisor

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type fakeClient struct {
	resp string
	err  error
}

func (f fakeClient) ChatCompletion(ctx context.Context, prompt string) (string, error) {
	return f.resp, f.err
}

var _ = Describe("Advisor", func() {
	It("returns suggestions", func() {
		a := New(fakeClient{resp: "refactor"})
		a.Run = func(ctx context.Context, name string, args ...string) ([]byte, error) { return []byte("ok"), nil }
		msg, err := a.Advise(context.Background())
		Expect(err).NotTo(HaveOccurred())
		Expect(msg).To(Equal("refactor"))
	})
})

func TestAdvisor(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Advisor Suite")
}
