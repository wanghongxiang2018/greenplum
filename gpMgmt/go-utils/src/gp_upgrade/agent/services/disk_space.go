package services

import (
	"context"

	pb "gp_upgrade/idl"

	"github.com/cloudfoundry/gosigar"
	"github.com/greenplum-db/gp-common-go-libs/gplog"
)

func (s *AgentServer) CheckDiskUsageOnAgents(ctx context.Context, in *pb.CheckDiskUsageRequestToAgent) (*pb.CheckDiskUsageReplyFromAgent, error) {
	gplog.Info("got a check disk command from the hub")
	diskUsage, err := s.GetDiskUsage()
	if err != nil {
		gplog.Error(err.Error())
		return nil, err
	}
	var listDiskUsages []*pb.FileSysUsage
	for k, v := range diskUsage {
		listDiskUsages = append(listDiskUsages, &pb.FileSysUsage{Filesystem: k, Usage: v})
	}
	return &pb.CheckDiskUsageReplyFromAgent{ListOfFileSysUsage: listDiskUsages}, nil
}

// diskUsage() wraps a pair of calls to the gosigar library.
// This is local repetition of the sys_utils function pointer pattern. If there was more than one of these,
// we would've refactored.
// "Adapted" from the gosigar usage example at https://github.com/cloudfoundry/gosigar/blob/master/examples/df.go
func diskUsage() (map[string]float64, error) {
	diskUsagePerFS := make(map[string]float64)
	fslist := sigar.FileSystemList{}
	err := fslist.Get()
	if err != nil {
		gplog.Error(err.Error())
		return nil, err
	}

	for _, fs := range fslist.List {
		dirName := fs.DirName

		usage := sigar.FileSystemUsage{}

		err = usage.Get(dirName)
		if err != nil {
			gplog.Error(err.Error())
			return nil, err
		}

		diskUsagePerFS[dirName] = usage.UsePercent()
	}
	return diskUsagePerFS, nil
}
