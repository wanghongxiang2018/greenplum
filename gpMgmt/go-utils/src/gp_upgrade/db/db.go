package db

import (
	"strconv"

	"gp_upgrade/utils"

	"github.com/greenplum-db/gp-common-go-libs/dbconn"
	_ "github.com/lib/pq" //_ import for the side effect of having postgres driver available
)

func NewDBConn(masterHost string, masterPort int, dbname string) *dbconn.DBConn {
	currentUser, _, _ := utils.GetUser()
	username := utils.TryEnv("PGUSER", currentUser)
	if dbname == "" {
		dbname = utils.TryEnv("PGDATABASE", "")
	}
	hostname, _ := utils.GetHost()
	if masterHost == "" {
		masterHost = utils.TryEnv("PGHOST", hostname)
	}
	if masterPort == 0 {
		masterPort, _ = strconv.Atoi(utils.TryEnv("PGPORT", "15432"))
	}

	return &dbconn.DBConn{
		ConnPool: nil,
		NumConns: 0,
		Driver:   dbconn.GPDBDriver{},
		User:     username,
		DBName:   dbname,
		Host:     masterHost,
		Port:     masterPort,
		Tx:       nil,
		Version:  dbconn.GPDBVersion{},
	}
}
