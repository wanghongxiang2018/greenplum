package services

import (
	"context"

	"errors"
	"fmt"
	"github.com/greenplum-db/gp-common-go-libs/gplog"
	"gp_upgrade/hub/upgradestatus"
	pb "gp_upgrade/idl"
	"path/filepath"
)

func (s *AgentServer) CheckConversionStatus(ctx context.Context, in *pb.CheckConversionStatusRequest) (*pb.CheckConversionStatusReply, error) {
	if len(in.GetSegments()) == 0 {
		return nil, errors.New("no segment information was passed to the agent")
	}
	format := "%s - DBID %d - CONTENT ID %d - %s - %s"

	var replies []string
	var master string
	for _, segment := range in.GetSegments() {
		conversionStatus := upgradestatus.NewPGUpgradeStatusChecker(
			filepath.Join(s.conf.StateDir, "pg_upgrade", fmt.Sprintf("seg-%d", segment.GetContent())),
			segment.GetDataDir(),
			s.commandExecer,
		)

		status, err := conversionStatus.GetStatus()
		if err != nil {
			gplog.Error("Unable to get status for segment %d conversion: %s", segment.GetContent(), err)
			return &pb.CheckConversionStatusReply{}, err
		}

		if segment.GetDbid() == 1 && segment.GetContent() == -1 {
			master = fmt.Sprintf(format, status.Status.String(), segment.GetDbid(), segment.GetContent(), "MASTER", in.GetHostname())
		} else {
			replies = append(replies, fmt.Sprintf(format, status.Status.String(), segment.GetDbid(), segment.GetContent(), "PRIMARY", in.GetHostname()))
		}
	}

	replies = append([]string{master}, replies...)

	return &pb.CheckConversionStatusReply{
		Statuses: replies,
	}, nil
}
