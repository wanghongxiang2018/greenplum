package services

import (
	"fmt"

	pb "gp_upgrade/idl"

	"golang.org/x/net/context"
)

func (h *HubClient) StatusConversion(ctx context.Context, in *pb.StatusConversionRequest) (*pb.StatusConversionReply, error) {
	conns, err := h.AgentConns()
	if err != nil {
		return &pb.StatusConversionReply{}, err
	}

	segments := h.segmentsByHost()

	var statuses []string
	for _, conn := range conns {
		var agentSegments []*pb.SegmentInfo
		for _, segment := range segments[conn.Hostname] {
			agentSegments = append(agentSegments, &pb.SegmentInfo{
				Content: int32(segment.Content),
				Dbid:    int32(segment.Dbid),
				DataDir: segment.Datadir,
			})
		}

		status, err := pb.NewAgentClient(conn.Conn).CheckConversionStatus(context.Background(), &pb.CheckConversionStatusRequest{
			Segments: agentSegments,
			Hostname: conn.Hostname,
		})
		if err != nil {
			return &pb.StatusConversionReply{}, fmt.Errorf("agent on host %s returned an error when checking conversion status: %s", conn.Hostname, err)
		}

		statuses = append(statuses, status.GetStatuses()...)
	}

	return &pb.StatusConversionReply{
		ConversionStatuses: statuses,
	}, nil
}
