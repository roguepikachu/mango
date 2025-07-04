package testmeta

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Extract", func() {
	It("extracts metadata from go and ginkgo tests", func() {
		dir := GinkgoT().TempDir()
		os.WriteFile(filepath.Join(dir, "foo_test.go"), []byte(`package foo
import "testing"
func TestFoo(t *testing.T){}
`), 0o644)
		os.WriteFile(filepath.Join(dir, "bar_test.go"), []byte(`package foo
import . "github.com/onsi/ginkgo/v2"
var _ = Describe("Bar", func(){It("works", func(){})})
`), 0o644)
		old, _ := os.Getwd()
		os.Chdir(dir)
		defer os.Chdir(old)

		meta, err := Extract()
		Expect(err).NotTo(HaveOccurred())
		Expect(meta).To(ContainElements(
			Metadata{Name: "TestFoo", File: "foo_test.go", Package: "foo"},
			Metadata{Name: "Bar", File: "bar_test.go", Package: "foo", Ginkgo: true},
			Metadata{Name: "works", File: "bar_test.go", Package: "foo", Ginkgo: true},
		))
	})
})

func TestParser(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Parser Suite")
}
