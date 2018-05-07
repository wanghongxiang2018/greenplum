package integrations_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gp_upgrade/hub/cluster"
	"gp_upgrade/hub/configutils"
	"gp_upgrade/hub/services"
	pb "gp_upgrade/idl"
	"gp_upgrade/testutils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

var _ = Describe("upgrade convert master", func() {
	var (
		dir           string
		hub           *services.HubClient
		mockAgent     *testutils.MockAgentServer
		commandExecer *testutils.FakeCommandExecer
		oldDataDir    string
		oldBinDir     string
		newDataDir    string
		newBinDir     string

		outChan chan []byte
		errChan chan error
	)

	BeforeEach(func() {
		var err error
		dir, err = ioutil.TempDir("", "")
		Expect(err).ToNot(HaveOccurred())

		oldDataDir, err = ioutil.TempDir("", "")
		Expect(err).ToNot(HaveOccurred())
		oldBinDir, err = ioutil.TempDir("", "")
		Expect(err).ToNot(HaveOccurred())

		newDataDir, err = ioutil.TempDir("", "")
		Expect(err).ToNot(HaveOccurred())
		newBinDir, err = ioutil.TempDir("", "")
		Expect(err).ToNot(HaveOccurred())

		oldConfig := `[{
			"dbid": 1,
			"port": 5432,
			"host": "localhost"
		}]`

		testutils.WriteOldConfig(dir, oldConfig)

		newConfig := `[{
			"dbid": 1,
			"port": 6432,
			"host": "localhost"
		}]`

		testutils.WriteNewConfig(dir, newConfig)

		port, err = testutils.GetOpenPort()
		Expect(err).ToNot(HaveOccurred())

		var agentPort int
		mockAgent, agentPort = testutils.NewMockAgentServer()

		conf := &services.HubConfig{
			CliToHubPort:   port,
			HubToAgentPort: agentPort,
			StateDir:       dir,
		}

		reader := configutils.NewReader()

		outChan = make(chan []byte, 10)
		errChan = make(chan error, 10)

		commandExecer = &testutils.FakeCommandExecer{}
		commandExecer.SetOutput(&testutils.FakeCommand{
			Out: outChan,
			Err: errChan,
		})

		hub = services.NewHub(&cluster.Pair{}, &reader, grpc.DialContext, commandExecer.Exec, conf)
		go hub.Start()
	})

	AfterEach(func() {
		hub.Stop()
		mockAgent.Stop()
		os.RemoveAll(dir)
		Expect(checkPortIsAvailable(port)).To(BeTrue())
	})

	It("updates status PENDING to RUNNING then to COMPLETE if successful", func() {
		mockAgent.StatusConversionResponse = &pb.CheckConversionStatusReply{
			Statuses: []string{},
		}

		Expect(runStatusUpgrade()).To(ContainSubstring("PENDING - Run pg_upgrade on master"))

		trigger := make(chan struct{}, 1)
		commandExecer.SetOutput(&testutils.FakeCommand{
			Out:     outChan,
			Err:     errChan,
			Trigger: trigger,
		})
		outChan <- []byte("pid1")

		wg := &sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer GinkgoRecover()

			Eventually(func() string {
				outChan <- []byte("pid1")
				return runStatusUpgrade()
			}).Should(ContainSubstring("RUNNING - Run pg_upgrade on master"))

			f, err := os.Create(filepath.Join(dir, "pg_upgrade", "1.done"))
			Expect(err).ToNot(HaveOccurred())
			f.Write([]byte("Upgrade complete\n")) //need for status upgrade validation
			f.Close()

			trigger <- struct{}{}
		}()

		upgradeConvertMasterSession := runCommand(
			"upgrade",
			"convert-master",
			"--old-datadir", oldDataDir,
			"--old-bindir", oldBinDir,
			"--new-datadir", newDataDir,
			"--new-bindir", newBinDir,
		)
		Eventually(upgradeConvertMasterSession).Should(Exit(0))
		wg.Wait()

		commandExecer.SetOutput(&testutils.FakeCommand{})

		allCalls := strings.Join(commandExecer.Calls(), "")
		Expect(allCalls).To(ContainSubstring(newBinDir + "/pg_upgrade"))

		Expect(runStatusUpgrade()).To(ContainSubstring("COMPLETE - Run pg_upgrade on master"))
	})

	It("updates status to FAILED if it fails to run", func() {
		mockAgent.StatusConversionResponse = &pb.CheckConversionStatusReply{
			Statuses: []string{},
		}

		Expect(runStatusUpgrade()).To(ContainSubstring("PENDING - Run pg_upgrade on master"))

		errChan <- errors.New("start failed")

		upgradeConvertMasterSession := runCommand(
			"upgrade",
			"convert-master",
			"--old-datadir", oldDataDir,
			"--old-bindir", oldBinDir,
			"--new-datadir", newDataDir,
			"--new-bindir", newBinDir,
		)
		Expect(upgradeConvertMasterSession).Should(Exit(1))

		Expect(runStatusUpgrade()).To(ContainSubstring("FAILED - Run pg_upgrade on master"))
	})

	It("fails if the --old-bindir or --new-bindir flags are missing", func() {
		prepareShutdownClustersSession := runCommand("upgrade", "convert-master")
		Expect(prepareShutdownClustersSession).Should(Exit(1))
		Expect(string(prepareShutdownClustersSession.Out.Contents())).To(Equal("Required flag(s) \"new-bindir\", \"new-datadir\", \"old-bindir\", \"old-datadir\" have/has not been set\n"))
	})
})
