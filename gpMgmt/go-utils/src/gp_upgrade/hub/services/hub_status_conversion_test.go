package services_test

import (
	"errors"

	"gp_upgrade/testutils"

	"google.golang.org/grpc"

	pb "gp_upgrade/idl"

	"gp_upgrade/hub/services"

	"gp_upgrade/hub/configutils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gp_upgrade/utils"
)

var _ = Describe("hub", func() {
	var (
		hubClient *services.HubClient
		agentA    *testutils.MockAgentServer
	)

	BeforeEach(func() {
		var port int
		agentA, port = testutils.NewMockAgentServer()

		segmentConfs := make(chan configutils.SegmentConfiguration, 1)
		reader := &testutils.SpyReader{
			Hostnames:             []string{"localhost", "localhost"},
			SegmentConfigurations: segmentConfs,
		}

		segmentConfs <- configutils.SegmentConfiguration{
			{
				Content:  0,
				Dbid:     2,
				Hostname: "localhost",
				Datadir:  "/first/data/dir",
			}, {
				Content:  1,
				Dbid:     3,
				Hostname: "localhost",
				Datadir:  "/second/data/dir",
			},
		}

		conf := &services.HubConfig{
			HubToAgentPort: port,
		}

		hubClient = services.NewHub(nil, reader, grpc.DialContext, nil, conf)
	})

	AfterEach(func() {
		utils.System = utils.InitializeSystemFunctions()
		agentA.Stop()
	})

	It("receives a conversion status from each agent and returns all as single message", func() {
		statusMessages := []string{"status", "status"}
		agentA.StatusConversionResponse = &pb.CheckConversionStatusReply{
			Statuses: statusMessages,
		}

		status, err := hubClient.StatusConversion(nil, &pb.StatusConversionRequest{})
		Expect(err).ToNot(HaveOccurred())

		Expect(status.GetConversionStatuses()).To(Equal([]string{"status", "status", "status", "status"}))
		Expect(agentA.StatusConversionRequest.GetHostname()).To(Equal("localhost"))
		Expect(agentA.StatusConversionRequest.GetSegments()).To(ConsistOf([]*pb.SegmentInfo{
			{
				Content: 0,
				Dbid:    2,
				DataDir: "/first/data/dir",
			},
			{
				Content: 1,
				Dbid:    3,
				DataDir: "/second/data/dir",
			},
		}))
	})

	It("returns an error when AgentConns returns an error", func() {
		agentA.Stop()

		_, err := hubClient.StatusConversion(nil, &pb.StatusConversionRequest{})
		Expect(err).To(HaveOccurred())
	})

	It("returns an error when Agent server returns an error", func() {
		agentA.Err <- errors.New("any error")

		_, err := hubClient.StatusConversion(nil, &pb.StatusConversionRequest{})
		Expect(err).To(HaveOccurred())
	})
})
