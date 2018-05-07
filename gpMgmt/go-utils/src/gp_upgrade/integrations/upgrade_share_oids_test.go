package integrations_test

import (
	"errors"
	"io/ioutil"
	"os"

	agentServices "gp_upgrade/agent/services"
	"gp_upgrade/hub/cluster"
	"gp_upgrade/hub/configutils"
	hubServices "gp_upgrade/hub/services"
	"gp_upgrade/testutils"

	"google.golang.org/grpc"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("upgrade share oids", func() {
	var (
		dir       string
		hub       *hubServices.HubClient
		agent     *agentServices.AgentServer
		hubExecer *testutils.FakeCommandExecer
		agentPort int

		outChan chan []byte
		errChan chan error
	)

	BeforeEach(func() {
		var err error
		dir, err = ioutil.TempDir("", "")
		Expect(err).ToNot(HaveOccurred())

		config := `[{
			"dbid": 1,
			"port": 5432,
			"host": "localhost"
		}]`

		testutils.WriteOldConfig(dir, config)
		testutils.WriteNewConfig(dir, config)

		agentPort, err = testutils.GetOpenPort()
		Expect(err).ToNot(HaveOccurred())

		agentConfig := agentServices.AgentConfig{
			Port:     agentPort,
			StateDir: dir,
		}

		agentExecer := &testutils.FakeCommandExecer{}
		agentExecer.SetOutput(&testutils.FakeCommand{})

		agent = agentServices.NewAgentServer(agentExecer.Exec, agentConfig)
		go agent.Start()

		port, err = testutils.GetOpenPort()
		Expect(err).ToNot(HaveOccurred())

		conf := &hubServices.HubConfig{
			CliToHubPort:   port,
			HubToAgentPort: agentPort,
			StateDir:       dir,
		}

		reader := configutils.NewReader()

		outChan = make(chan []byte, 10)
		errChan = make(chan error, 10)
		hubExecer = &testutils.FakeCommandExecer{}
		hubExecer.SetOutput(&testutils.FakeCommand{
			Out: outChan,
			Err: errChan,
		})

		hub = hubServices.NewHub(&cluster.Pair{}, &reader, grpc.DialContext, hubExecer.Exec, conf)
		go hub.Start()
	})

	AfterEach(func() {
		hub.Stop()
		agent.Stop()

		os.RemoveAll(dir)

		Expect(checkPortIsAvailable(port)).To(BeTrue())
		Expect(checkPortIsAvailable(agentPort)).To(BeTrue())
	})

	It("updates status PENDING to RUNNING then to COMPLETE if successful", func() {
		Expect(runStatusUpgrade()).To(ContainSubstring("PENDING - Copy OID files from master to segments"))

		trigger := make(chan struct{}, 1)
		hubExecer.SetTrigger(trigger)

		upgradeShareOidsSession := runCommand("upgrade", "share-oids")
		Eventually(upgradeShareOidsSession).Should(Exit(0))

		Eventually(runStatusUpgrade()).Should(ContainSubstring("RUNNING - Copy OID files from master to segments"))
		trigger <- struct{}{}

		Expect(hubExecer.Calls()[0]).To(ContainSubstring("rsync"))

		Expect(runStatusUpgrade()).To(ContainSubstring("COMPLETE - Copy OID files from master to segments"))
	})

	It("updates status to FAILED if it fails to run", func() {
		Expect(runStatusUpgrade()).To(ContainSubstring("PENDING - Copy OID files from master to segments"))

		errChan <- errors.New("fake test error, share oid failed to send files")

		upgradeShareOidsSession := runCommand("upgrade", "share-oids")
		Eventually(upgradeShareOidsSession).Should(Exit(0))

		Eventually(runStatusUpgrade()).Should(ContainSubstring("FAILED - Copy OID files from master to segments"))
	})
})
