package integrations_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("check object-count", func() {
	It("fails if the --master-host flag is missing", func() {
		checkObjectCountSession := runCommand("check", "object-count")
		Expect(checkObjectCountSession).Should(Exit(1))
		Expect(string(checkObjectCountSession.Out.Contents())).To(Equal("Required flag(s) \"master-host\" have/has not been set\n"))
	})
})
