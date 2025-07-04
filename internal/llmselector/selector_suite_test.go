package llmselector

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestLLMSelector(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "LLMSelector Suite")
}
