package diff

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("changedFunctions", func() {
	It("detects modified functions", func() {
		dir := GinkgoT().TempDir()
		file := filepath.Join(dir, "sample.go")
		os.WriteFile(file, []byte(`package sample
func A() {}

func B() {}
`), 0o644)

		funcs, err := changedFunctions(file, []int{2})
		Expect(err).NotTo(HaveOccurred())
		Expect(funcs).To(ConsistOf("A"))

		funcs, err = changedFunctions(file, []int{4})
		Expect(err).NotTo(HaveOccurred())
		Expect(funcs).To(ConsistOf("B"))

		funcs, err = changedFunctions(file, []int{2, 4})
		Expect(err).NotTo(HaveOccurred())
		Expect(funcs).To(ConsistOf("A", "B"))
	})
})

func TestDiff(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Diff Suite")
}
