package services

import (
	"encoding/json"
	"fmt"
	"io"
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

type VersionNum struct {
	GpServerVersionNum int `db:"gp_server_version_num"`
}

var CONFIGQUERY5 = `SELECT
	s.content,
	s.hostname,
	e.fselocation as datadir,
	s.dbid,
	s.preferred_role,
	s.role,
	s.port
	FROM gp_segment_configuration s
	JOIN pg_filespace_entry e ON s.dbid = e.fsedbid
	JOIN pg_filespace f ON e.fsefsoid = f.oid
	WHERE s.role = 'p' AND f.fsname = 'pg_system'
	ORDER BY s.content;`

var CONFIGQUERY6 = `SELECT
	content,
	hostname,
	datadir,
	dbid,
	preferred_role,
	role,
	port
	FROM gp_segment_configuration
	WHERE role = 'p'
	ORDER BY content;`

func (h *HubClient) CheckConfig(ctx context.Context,
	in *pb.CheckConfigRequest) (*pb.CheckConfigReply, error) {
	gplog.Info("starting CheckConfig()")

	dbConnector := db.NewDBConn("localhost", int(in.DbPort), "template1")
	defer dbConnector.Close()
	err := dbConnector.Connect(1)
	if err != nil {
		gplog.Error(err.Error())
		return &pb.CheckConfigReply{}, utils.DatabaseConnectionError{Parent: err}
	}
	dbConnector.Version.Initialize(dbConnector)

	err = SaveOldClusterConfigAndVersion(dbConnector, h.conf.StateDir)
	if err != nil {
		gplog.Error(err.Error())
		return &pb.CheckConfigReply{}, err
	}

	successReply := &pb.CheckConfigReply{ConfigStatus: "All good"}

	return successReply, nil
}

func SaveOldClusterConfigAndVersion(dbConnector *dbconn.DBConn, stateDir string) error {
	err := os.MkdirAll(stateDir, 0700)
	if err != nil {
		return err
	}

	configQuery := CONFIGQUERY6
	if dbConnector.Version.Before("6") {
		configQuery = CONFIGQUERY5
	}

	configFile := configutils.GetConfigFilePath(stateDir)
	configFileHandle, err := operating.System.OpenFileWrite(configFile, os.O_CREATE|os.O_WRONLY, 0700)
	if err != nil {
		errMsg := fmt.Sprintf("Unable to write to config file %s. Err: %s", configFile, err.Error())
		return errors.New(errMsg)
	}

	segConfig := make(configutils.SegmentConfiguration, 0)
	err = dbConnector.Select(&segConfig, configQuery)
	if err != nil {
		errMsg := fmt.Sprintf("Unable to execute query %s. Err: %s", configQuery, err.Error())
		return errors.New(errMsg)
	}

	err = SaveQueryResultToJSON(&segConfig, configFileHandle)
	if err != nil {
		return err
	}

	versionFile := configutils.GetVersionFilePath(stateDir)
	versionFileHandle, err := operating.System.OpenFileWrite(versionFile, os.O_CREATE|os.O_WRONLY, 0700)
	if err != nil {
		errMsg := fmt.Sprintf("Unable to write to version file %s. Err: %s", versionFile, err.Error())
		return errors.New(errMsg)
	}

	versionQuery := `show gp_server_version_num`
	versionNum := VersionNum{}
	err = dbConnector.Get(&versionNum, versionQuery)
	if err != nil {
		errMsg := fmt.Sprintf("Unable to execute query %s. Err: %s", versionQuery, err.Error())
		return errors.New(errMsg)
	}
	versionSlice := make([]VersionNum, 1)
	versionSlice[0] = versionNum

	err = SaveQueryResultToJSON(&versionSlice, versionFileHandle)
	if err != nil {
		return err
	}

	return nil
}

// public for testing purposes
func SaveQueryResultToJSON(structure interface{}, fileHandle io.WriteCloser) error {
	byteArr, err := json.Marshal(structure)
	if err != nil {
		errMsg := fmt.Sprintf("Unable to marshal query results to JSON. Err: %s", err.Error())
		return errors.New(errMsg)
	}

	_, err = fileHandle.Write(byteArr)
	if err != nil {
		errMsg := fmt.Sprintf("Unable to write query results to file. Err: %s", err.Error())
		return errors.New(errMsg)
	}

	return nil
}
