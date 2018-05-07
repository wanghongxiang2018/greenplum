package services

import (
	"fmt"
	"os"

	"gp_upgrade/db"
	"gp_upgrade/hub/configutils"
	pb "gp_upgrade/idl"
	"gp_upgrade/utils"

	"github.com/greenplum-db/gp-common-go-libs/dbconn"
	"github.com/greenplum-db/gp-common-go-libs/gplog"
	"github.com/greenplum-db/gp-common-go-libs/operating"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

func SaveTargetClusterConfig(dbConnector *dbconn.DBConn, stateDir string) error {
	segConfig := make(configutils.SegmentConfiguration, 0)
	var configQuery string

	configQuery = CONFIGQUERY6
	if dbConnector.Version.Before("6") {
		configQuery = CONFIGQUERY5
	}

	configFile, err := operating.System.OpenFileWrite(configutils.GetNewClusterConfigFilePath(stateDir), os.O_CREATE|os.O_WRONLY, 0700)
	if err != nil {
		errMsg := fmt.Sprintf("Could not open new config file for writing. Err: %s", err.Error())
		return errors.New(errMsg)
	}

	err = dbConnector.Select(&segConfig, configQuery)
	if err != nil {
		errMsg := fmt.Sprintf("Unable to execute query %s. Err: %s", configQuery, err.Error())
		return errors.New(errMsg)
	}

	err = SaveQueryResultToJSON(&segConfig, configFile)
	if err != nil {
		return err
	}

	return nil
}

func (h *HubClient) PrepareInitCluster(ctx context.Context, in *pb.PrepareInitClusterRequest) (*pb.PrepareInitClusterReply, error) {
	gplog.Info("starting PrepareInitCluster()")

	dbConnector := db.NewDBConn("localhost", int(in.DbPort), "template1")
	defer dbConnector.Close()
	err := dbConnector.Connect(1)
	if err != nil {
		gplog.Error(err.Error())
		return &pb.PrepareInitClusterReply{}, utils.DatabaseConnectionError{Parent: err}
	}
	dbConnector.Version.Initialize(dbConnector)

	err = SaveTargetClusterConfig(dbConnector, h.conf.StateDir)
	if err != nil {
		gplog.Error(err.Error())
		return &pb.PrepareInitClusterReply{}, err
	}

	return &pb.PrepareInitClusterReply{}, nil
}
