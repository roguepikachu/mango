package query

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

var _ = Describe("Service", func() {
	It("answers questions", func() {
		s := New(fakeClient{resp: "answer"})
		msg, err := s.Ask(context.Background(), "which", []testmeta.Metadata{{Name: "A"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(msg).To(Equal("answer"))
	})
})

func TestService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Query Suite")
}
