package services

import (
	"context"

	pb "gp_upgrade/idl"

	"github.com/greenplum-db/gp-common-go-libs/gplog"
)

func (s *AgentServer) PingAgents(ctx context.Context, in *pb.PingAgentsRequest) (*pb.PingAgentsReply, error) {
	gplog.Info("Successfully pinged agent")
	return &pb.PingAgentsReply{}, nil
}
