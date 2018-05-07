package services_test

import (
	pb "gp_upgrade/idl"
	mockpb "gp_upgrade/mock_idl"

	"gp_upgrade/hub/configutils"
	"gp_upgrade/hub/services"

	"github.com/golang/mock/gomock"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"gp_upgrade/utils"
)

var _ = Describe("object count tests", func() {
	var (
		client *mockpb.MockAgentClient
		ctrl   *gomock.Controller
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		client = mockpb.NewMockAgentClient(ctrl)
	})

	AfterEach(func() {
		utils.System = utils.InitializeSystemFunctions()
		ctrl.Finish()
	})

	Describe("GetDiskUsageFromSegmentHosts", func() {
		It("returns err msg when unable to call CheckDiskUsageOnAgents on segment host", func() {
			var clients []configutils.ClientAndHostname

			client.EXPECT().CheckDiskUsageOnAgents(
				gomock.Any(),
				&pb.CheckDiskUsageRequestToAgent{},
			).Return(&pb.CheckDiskUsageReplyFromAgent{}, errors.New("couldn't connect to hub"))
			clients = append(clients, configutils.ClientAndHostname{Client: client, Hostname: "doesnotexist"})

			messages := services.GetDiskUsageFromSegmentHosts(clients)
			Expect(len(messages)).To(Equal(1))
			Expect(messages[0]).To(ContainSubstring("Could not get disk usage from: "))
		})

		It("lists filesystems above usage threshold", func() {
			var clients []configutils.ClientAndHostname

			var expectedFilesystemsUsage []*pb.FileSysUsage
			expectedFilesystemsUsage = append(expectedFilesystemsUsage, &pb.FileSysUsage{Filesystem: "first filesystem", Usage: 90.4})
			expectedFilesystemsUsage = append(expectedFilesystemsUsage, &pb.FileSysUsage{Filesystem: "/second/filesystem", Usage: 24.2})

			client.EXPECT().CheckDiskUsageOnAgents(
				gomock.Any(),
				&pb.CheckDiskUsageRequestToAgent{},
			).Return(&pb.CheckDiskUsageReplyFromAgent{ListOfFileSysUsage: expectedFilesystemsUsage}, nil)
			clients = append(clients, configutils.ClientAndHostname{Client: client, Hostname: "doesnotexist"})

			messages := services.GetDiskUsageFromSegmentHosts(clients)
			Expect(len(messages)).To(Equal(1))
			Expect(messages[0]).To(ContainSubstring("diskspace check - doesnotexist - WARNING first filesystem 90.4 use"))
		})

		It("lists hosts for which all filesystems are below usage threshold", func() {
			var clients []configutils.ClientAndHostname

			var expectedFilesystemsUsage []*pb.FileSysUsage
			expectedFilesystemsUsage = append(expectedFilesystemsUsage, &pb.FileSysUsage{Filesystem: "first filesystem", Usage: 70.4})
			expectedFilesystemsUsage = append(expectedFilesystemsUsage, &pb.FileSysUsage{Filesystem: "/second/filesystem", Usage: 24.2})

			client.EXPECT().CheckDiskUsageOnAgents(
				gomock.Any(),
				&pb.CheckDiskUsageRequestToAgent{},
			).Return(&pb.CheckDiskUsageReplyFromAgent{ListOfFileSysUsage: expectedFilesystemsUsage}, nil)
			clients = append(clients, configutils.ClientAndHostname{Client: client, Hostname: "doesnotexist"})

			messages := services.GetDiskUsageFromSegmentHosts(clients)
			Expect(len(messages)).To(Equal(1))
			Expect(messages[0]).To(ContainSubstring("diskspace check - doesnotexist - OK"))
		})
	})
})
