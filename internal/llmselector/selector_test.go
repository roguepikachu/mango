package llmselector

import (
	"encoding/json"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("parseResponse", func() {
	It("parses json array", func() {
		names, err := parseResponse(`["TestFoo","TestBar"]`)
		Expect(err).NotTo(HaveOccurred())
		Expect(names).To(ConsistOf("TestFoo", "TestBar"))
	})

	It("parses newline list", func() {
		names, err := parseResponse("TestFoo\nTestBar\n")
		Expect(err).NotTo(HaveOccurred())
		Expect(names).To(ConsistOf("TestFoo", "TestBar"))
	})

	It("returns error on invalid", func() {
		_, err := parseResponse(" ")
		Expect(err).To(HaveOccurred())
	})
})

func FuzzParseResponse(f *testing.F) {
	f.Add(`["A"]`)
	f.Fuzz(func(t *testing.T, s string) {
		var arr []string
		json.Unmarshal([]byte(s), &arr)
		parseResponse(s)
	})
}
