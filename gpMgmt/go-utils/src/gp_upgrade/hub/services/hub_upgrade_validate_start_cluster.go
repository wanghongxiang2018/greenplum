package services

import (
	"fmt"
	"os"

	pb "gp_upgrade/idl"

	"gp_upgrade/hub/upgradestatus"

	"github.com/greenplum-db/gp-common-go-libs/gplog"
	"golang.org/x/net/context"
)

func (h *HubClient) UpgradeValidateStartCluster(ctx context.Context,
	in *pb.UpgradeValidateStartClusterRequest) (*pb.UpgradeValidateStartClusterReply, error) {
	gplog.Info("Started processing validate-start-cluster request")

	go h.startNewCluster(in.NewBinDir, in.NewDataDir)

	return &pb.UpgradeValidateStartClusterReply{}, nil
}

func (h *HubClient) startNewCluster(newBinDir string, newDataDir string) {
	gplog.Error(h.conf.StateDir)
	c := upgradestatus.NewChecklistManager(h.conf.StateDir)
	err := c.ResetStateDir("validate-start-cluster")
	if err != nil {
		gplog.Error("failed to reset the state dir for validate-start-cluster")

		return
	}

	err = c.MarkInProgress("validate-start-cluster")
	if err != nil {
		gplog.Error("failed to record in-progress for validate-start-cluster")

		return
	}

	commandArgs := fmt.Sprintf("PYTHONPATH=%s && %s/gpstart -a -d %s", os.Getenv("PYTHONPATH"), newBinDir, newDataDir)
	_, err = h.commandExecer("bash", "-c", commandArgs).Output()
	if err != nil {
		gplog.Error(err.Error())
		cmErr := c.MarkFailed("validate-start-cluster")
		if cmErr != nil {
			gplog.Error("failed to record failed for validate-start-cluster")
		}

		return
	}

	err = c.MarkComplete("validate-start-cluster")
	if err != nil {
		gplog.Error("failed to record completed for validate-start-cluster")
		return
	}

	return
}
