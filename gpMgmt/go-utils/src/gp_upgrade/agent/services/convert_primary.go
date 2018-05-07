package services

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	pb "gp_upgrade/idl"
	"gp_upgrade/utils"

	"github.com/greenplum-db/gp-common-go-libs/gplog"
)

func (s *AgentServer) UpgradeConvertPrimarySegments(ctx context.Context, in *pb.UpgradeConvertPrimarySegmentsRequest) (*pb.UpgradeConvertPrimarySegmentsReply, error) {
	gplog.Info("got a request to convert primary from the hub")

	filename := "pg_upgrade_dump_*_oids.sql"
	shareOIDfilePath := filepath.Join(s.conf.StateDir, "pg_upgrade", filename)
	oidFiles, err := utils.System.FilePathGlob(shareOIDfilePath)
	if err != nil {
		gplog.Error("ls OID files failed. Err: %v", err)
		return &pb.UpgradeConvertPrimarySegmentsReply{}, err
	}
	//len(nil) = 0
	if len(oidFiles) == 0 {
		gplog.Error("Share OID files do not exist. Err: %s", shareOIDfilePath)
		return &pb.UpgradeConvertPrimarySegmentsReply{}, errors.New("No OID files found")
	}

	for _, segment := range in.DataDirPairs {
		pathToSegment := filepath.Join(s.conf.StateDir, "pg_upgrade", fmt.Sprintf("seg-%d", segment.Content))
		err := utils.System.MkdirAll(pathToSegment, 0700)
		if err != nil {
			gplog.Error("Could not create segment directory. Err: %v", err)
			return &pb.UpgradeConvertPrimarySegmentsReply{}, err
		}

		for _, oidFile := range oidFiles {
			out, err := s.commandExecer("cp", oidFile, pathToSegment).CombinedOutput()
			if err != nil {
				gplog.Error("Failed to copy OID files for segment %d. Output: %s", segment.Content, string(out))
				return &pb.UpgradeConvertPrimarySegmentsReply{}, err
			}
		}

		convertPrimaryArgs := fmt.Sprintf("cd %s && nohup %s --old-bindir=%s --old-datadir=%s --new-bindir=%s --new-datadir=%s --old-port=%d --new-port=%d --progress",
			pathToSegment, in.NewBinDir+"/pg_upgrade", in.OldBinDir, segment.OldDataDir, in.NewBinDir, segment.NewDataDir, segment.OldPort, segment.NewPort)

		convertPrimaryCmd := s.commandExecer("bash", "-c", convertPrimaryArgs)

		err = convertPrimaryCmd.Start()
		if err != nil {
			gplog.Error("An error occurred: %v", err)
			return &pb.UpgradeConvertPrimarySegmentsReply{}, err
		}
	}

	return &pb.UpgradeConvertPrimarySegmentsReply{}, nil
}
