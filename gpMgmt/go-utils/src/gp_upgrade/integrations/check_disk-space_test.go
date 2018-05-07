package integrations_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("check disk-space", func() {
	It("fails if the --master-host flag is missing", func() {
		checkDiskSpaceSession := runCommand("check", "disk-space")
		Expect(checkDiskSpaceSession).Should(Exit(1))
		Expect(string(checkDiskSpaceSession.Out.Contents())).To(Equal("Required flag(s) \"master-host\" have/has not been set\n"))
	})
})
