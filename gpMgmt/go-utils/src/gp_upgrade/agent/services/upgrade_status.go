package services

import (
	"context"

	pb "gp_upgrade/idl"

	"github.com/greenplum-db/gp-common-go-libs/gplog"
)

func (s *AgentServer) CheckUpgradeStatus(ctx context.Context, in *pb.CheckUpgradeStatusRequest) (*pb.CheckUpgradeStatusReply, error) {
	cmd := "ps auxx | grep pg_upgrade"

	output, err := s.commandExecer("bash", "-c", cmd).Output()
	if err != nil {
		gplog.Error(err.Error())
		return nil, err
	}
	return &pb.CheckUpgradeStatusReply{ProcessList: string(output)}, nil
}
