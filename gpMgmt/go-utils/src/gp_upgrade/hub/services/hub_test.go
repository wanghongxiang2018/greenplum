package services_test

import (
	"gp_upgrade/hub/services"
	"gp_upgrade/testutils"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gp_upgrade/utils"
)

var _ = Describe("HubClient", func() {
	var (
		reader *testutils.SpyReader
		agentA *testutils.MockAgentServer
		port   int
	)

	BeforeEach(func() {
		reader = &testutils.SpyReader{}
		agentA, port = testutils.NewMockAgentServer()
	})

	AfterEach(func() {
		utils.System = utils.InitializeSystemFunctions()
		agentA.Stop()
	})

	It("closes open connections when shutting down", func(done Done) {
		defer close(done)
		reader.Hostnames = []string{"localhost"}
		hub := services.NewHub(nil, reader, grpc.DialContext, nil, &services.HubConfig{
			HubToAgentPort: port,
		})
		go hub.Start()

		By("creating connections")
		conns, err := hub.AgentConns()
		Expect(err).ToNot(HaveOccurred())

		Eventually(func() connectivity.State { return conns[0].Conn.GetState() }).Should(Equal(connectivity.Ready))

		By("closing the connections")
		hub.Stop()
		Expect(err).ToNot(HaveOccurred())

		Eventually(func() connectivity.State { return conns[0].Conn.GetState() }).Should(Equal(connectivity.Shutdown))
	})

	It("retrieves the agent connections from the config file reader", func() {
		reader.Hostnames = []string{"localhost", "localhost"}
		hub := services.NewHub(nil, reader, grpc.DialContext, nil, &services.HubConfig{
			HubToAgentPort: port,
		})

		conns, err := hub.AgentConns()
		Expect(err).ToNot(HaveOccurred())

		Eventually(func() connectivity.State { return conns[0].Conn.GetState() }).Should(Equal(connectivity.Ready))
		Expect(conns[0].Hostname).To(Equal("localhost"))
		Eventually(func() connectivity.State { return conns[1].Conn.GetState() }).Should(Equal(connectivity.Ready))
		Expect(conns[1].Hostname).To(Equal("localhost"))
	})

	It("saves grpc connections for future calls", func() {
		reader.Hostnames = []string{"localhost"}

		hub := services.NewHub(nil, reader, grpc.DialContext, nil, &services.HubConfig{
			HubToAgentPort: port,
		})

		newConns, err := hub.AgentConns()
		Expect(err).ToNot(HaveOccurred())
		Expect(newConns).To(HaveLen(1))

		savedConns, err := hub.AgentConns()
		Expect(err).ToNot(HaveOccurred())
		Expect(savedConns).To(HaveLen(1))

		Expect(newConns[0]).To(Equal(savedConns[0]))
	})

	It("returns an error if any connections have non-ready states", func() {
		reader.Hostnames = []string{"localhost"}
		hub := services.NewHub(nil, reader, grpc.DialContext, nil, &services.HubConfig{
			HubToAgentPort: port,
		})

		conns, err := hub.AgentConns()
		Expect(err).ToNot(HaveOccurred())
		Expect(conns).To(HaveLen(1))

		agentA.Stop()

		Eventually(func() connectivity.State { return conns[0].Conn.GetState() }).Should(Equal(connectivity.TransientFailure))

		_, err = hub.AgentConns()
		Expect(err).To(HaveOccurred())
	})

	It("returns an error if any connections have non-ready states when first dialing", func() {
		reader.Hostnames = []string{"localhost"}
		hub := services.NewHub(nil, reader, grpc.DialContext, nil, &services.HubConfig{
			HubToAgentPort: port,
		})

		agentA.Stop()

		_, err := hub.AgentConns()
		Expect(err).To(HaveOccurred())
	})

	It("returns an error if the grpc dialer to the agent throws an error", func() {
		agentA.Stop()

		reader.Hostnames = []string{"example"}
		hub := services.NewHub(nil, reader, grpc.DialContext, nil, &services.HubConfig{
			HubToAgentPort: port,
		})

		_, err := hub.AgentConns()
		Expect(err).To(HaveOccurred())
	})

	It("returns an error if the config reader fails", func() {
		reader.Err = errors.New("error occurred while getting hostnames")
		hub := services.NewHub(nil, reader, nil, nil, &services.HubConfig{})

		_, err := hub.AgentConns()
		Expect(err).To(HaveOccurred())
	})
})
