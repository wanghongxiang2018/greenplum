package integrations_test

import (
	"io/ioutil"
	"os"
	"strings"
	"sync"

	"gp_upgrade/hub/cluster"
	"gp_upgrade/hub/services"
	pb "gp_upgrade/idl"
	"gp_upgrade/testutils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
	"google.golang.org/grpc"
)

var _ = Describe("check", func() {
	var (
		dir           string
		hub           *services.HubClient
		mockAgent     *testutils.MockAgentServer
		commandExecer *testutils.FakeCommandExecer
		outChan       chan []byte
		errChan       chan error
	)

	BeforeEach(func() {
		var err error
		dir, err = ioutil.TempDir("", "")
		Expect(err).ToNot(HaveOccurred())

		config := `[{
			"content": 2,
			"dbid": 7,
			"hostname": "localhost"
		}]`
		testutils.WriteOldConfig(dir, config)
		testutils.WriteNewConfig(dir, config)

		port, err = testutils.GetOpenPort()
		Expect(err).ToNot(HaveOccurred())

		var agentPort int
		mockAgent, agentPort = testutils.NewMockAgentServer()

		conf := &services.HubConfig{
			CliToHubPort:   port,
			HubToAgentPort: agentPort,
			StateDir:       dir,
		}
		reader := &testutils.SpyReader{
			Hostnames: []string{"localhost"},
		}

		outChan = make(chan []byte, 2)
		errChan = make(chan error, 2)

		commandExecer = &testutils.FakeCommandExecer{}
		commandExecer.SetOutput(&testutils.FakeCommand{
			Out: outChan,
			Err: errChan,
		})

		hub = services.NewHub(&cluster.Pair{}, reader, grpc.DialContext, commandExecer.Exec, conf)
		go hub.Start()
	})

	AfterEach(func() {
		hub.Stop()
		mockAgent.Stop()
		os.RemoveAll(dir)
	})

	// `gp_upgrade check seginstall` verifies that the user has installed the software on all hosts
	// As a single-node check, this test verifies the mechanics of the check, but would typically succeed.
	// The implementation, however, uses the gp_upgrade_agent binary to verify installation. In real life,
	// all the binaries, gp_upgrade_hub and gp_upgrade_agent included, would be alongside each other.
	// But in our integration tests' context, only the necessary Golang code is compiled, and Ginkgo's default
	// is to compile gp_upgrade_hub and gp_upgrade_agent in separate directories. As such, this test depends on the
	// setup in `integrations_suite_test.go` to replicate the real-world scenario of "install binaries side-by-side".
	//
	// TODO: This test might be interesting to run multi-node; for that, figure out how "installation" should be done
	Describe("seginstall", func() {
		It("updates status PENDING to RUNNING then to COMPLETE if successful", func() {
			mockAgent.StatusConversionResponse = &pb.CheckConversionStatusReply{
				Statuses: []string{},
			}

			Expect(runStatusUpgrade()).To(ContainSubstring("PENDING - Install binaries on segments"))

			trigger := make(chan struct{}, 2)
			commandExecer.SetTrigger(trigger)

			wg := &sync.WaitGroup{}
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer GinkgoRecover()

				Eventually(runStatusUpgrade).Should(ContainSubstring("RUNNING - Install binaries on segments"))
				trigger <- struct{}{}
			}()

			checkSeginstallSession := runCommand("check", "seginstall", "--master-host", "localhost")
			Eventually(checkSeginstallSession).Should(Exit(0))
			wg.Wait()

			Expect(commandExecer.Command()).To(Equal("ssh"))
			Expect(strings.Join(commandExecer.Args(), "")).To(ContainSubstring("ls"))
			Eventually(runStatusUpgrade).Should(ContainSubstring("COMPLETE - Install binaries on segments"))
		})
	})

	It("fails if the --master-host flag is missing", func() {
		checkSeginstallSession := runCommand("check", "seginstall")
		Expect(checkSeginstallSession).Should(Exit(1))
		Expect(string(checkSeginstallSession.Out.Contents())).To(Equal("Required flag(s) \"master-host\" have/has not been set\n"))
	})
})
