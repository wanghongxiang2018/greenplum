package services

import (
	pb "gp_upgrade/idl"

	"github.com/greenplum-db/gp-common-go-libs/gplog"
	"golang.org/x/net/context"
)

func (h *HubClient) Ping(ctx context.Context, in *pb.PingRequest) (*pb.PingReply, error) {
	gplog.Info("starting Ping")
	return &pb.PingReply{}, nil
}
