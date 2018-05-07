package services_test

import (
	"context"

	"gp_upgrade/agent/services"
	"gp_upgrade/testutils"
	"gp_upgrade/utils"

	"github.com/greenplum-db/gp-common-go-libs/testhelper"
	"github.com/onsi/gomega/gbytes"
	"github.com/pkg/errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CommandListener", func() {
	var (
		testLogFile   *gbytes.Buffer
		agent         *services.AgentServer
		commandExecer *testutils.FakeCommandExecer
		errChan       chan error
		outChan       chan []byte
	)

	BeforeEach(func() {
		_, _, testLogFile = testhelper.SetupTestLogger()

		errChan = make(chan error, 2)
		outChan = make(chan []byte, 2)

		commandExecer = &testutils.FakeCommandExecer{}
		commandExecer.SetOutput(&testutils.FakeCommand{
			Err: errChan,
			Out: outChan,
		})

		agent = services.NewAgentServer(commandExecer.Exec, services.AgentConfig{})
	})

	AfterEach(func() {
		//any mocking of utils.System function pointers should be reset by calling InitializeSystemFunctions
		utils.System = utils.InitializeSystemFunctions()
	})

	It("returns the shell command output", func() {
		outChan <- []byte("shell command output")

		resp, err := agent.CheckUpgradeStatus(context.TODO(), nil)
		Expect(err).ToNot(HaveOccurred())

		Expect(resp.ProcessList).To(Equal("shell command output"))
	})

	It("returns only err if anything is reported as an error", func() {
		errChan <- errors.New("couldn't find bash")

		resp, err := agent.CheckUpgradeStatus(context.TODO(), nil)
		Expect(err).To(HaveOccurred())

		Expect(resp).To(BeNil())
		Expect(testLogFile.Contents()).To(ContainSubstring("couldn't find bash"))
	})
})
