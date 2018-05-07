package services

import (
	"gp_upgrade/hub/configutils"
	pb "gp_upgrade/idl"

	"time"

	"github.com/greenplum-db/gp-common-go-libs/gplog"
	"golang.org/x/net/context"
)

type PingerManager struct {
	RPCClients       []configutils.ClientAndHostname
	NumRetries       int
	PauseBeforeRetry time.Duration
}

func NewPingerManager(baseDir string, t time.Duration) *PingerManager {
	rpcClients, err := configutils.GetClients(baseDir)
	if err != nil {
		return &PingerManager{}
	}
	return &PingerManager{rpcClients, 10, t}
}

func (agent *PingerManager) PingPollAgents() error {
	var err error
	done := false
	for i := 0; i < 10; i++ {
		gplog.Info("Pinging agents...")
		err = agent.PingAllAgents()
		if err == nil {
			done = true
			break
		}
		time.Sleep(agent.PauseBeforeRetry)
	}
	if !done {
		gplog.Info("Reached ping timeout")
	}
	return err
}

func (agent *PingerManager) PingAllAgents() error {
	//TODO: ping all the agents in parallel?
	for i := 0; i < len(agent.RPCClients); i++ {
		_, err := agent.RPCClients[i].Client.PingAgents(context.Background(), &pb.PingAgentsRequest{})
		if err != nil {
			gplog.Error("Not all agents on the segment hosts are running.")
			return err
		}
	}

	return nil
}
