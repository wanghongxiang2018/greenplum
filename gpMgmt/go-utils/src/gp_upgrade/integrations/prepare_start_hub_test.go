package integrations_test

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Start Hub", func() {
	var (
		dir string
	)

	BeforeEach(func() {
		killCommand := exec.Command("pkill", "-9", "gp_upgrade_hub")
		Start(killCommand, GinkgoWriter, GinkgoWriter)

		var err error
		dir, err = ioutil.TempDir("", "")
		Expect(err).ToNot(HaveOccurred())

		Expect(checkPortIsAvailable(port)).To(BeTrue())
	})

	AfterEach(func() {
		killCommand := exec.Command("pkill", "-9", "gp_upgrade_hub")
		Start(killCommand, GinkgoWriter, GinkgoWriter)

		Expect(checkPortIsAvailable(port)).To(BeTrue())
	})

	It("finds the right hub binary and starts a daemonized process", func() {
		gpUpgradeSession := runCommand("prepare", "start-hub")
		Eventually(gpUpgradeSession).Should(Exit(0))

		verificationCmd := exec.Command("bash", "-c", `ps -ef | grep -Gq "[g]p_upgrade_hub$"`)
		verificationSession, err := Start(verificationCmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(verificationSession).Should(Exit(0))
	})

	It("does not return an error if a non-gp_upgrade_hub process with gp_upgrade_hub in the name is running", func() {
		path := filepath.Join(dir, "gp_upgrade_hub_test_log")
		f, err := os.Create(path)
		Expect(err).ToNot(HaveOccurred())
		f.Close()

		tailCmd := exec.Command("tail", "-f", path)
		tailSession, err := Start(tailCmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		defer tailSession.Terminate()

		firstSession := runCommand("prepare", "start-hub")
		Expect(string(firstSession.Err.Contents())).Should(Equal(""))
		Eventually(firstSession).Should(Exit(0))
	})

	It("returns an error if gp_upgrade_hub is already running", func() {
		firstSession := runCommand("prepare", "start-hub")
		Eventually(firstSession).Should(Exit(0))
		//second start should return error
		secondSession := runCommand("prepare", "start-hub")
		Eventually(secondSession).Should(Exit(1))
	})

	It("returns an error if gp_upgrade_hub is not on the path", func() {
		origPath := os.Getenv("PATH")
		os.Setenv("PATH", "")
		gpUpgradeSession := runCommand("prepare", "start-hub")
		Eventually(gpUpgradeSession).ShouldNot(Exit(0))
		os.Setenv("PATH", origPath)
	})
})
