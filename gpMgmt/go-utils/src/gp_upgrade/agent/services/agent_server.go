package services

import (
	"net"
	"strconv"
	"sync"

	"gp_upgrade/helpers"
	pb "gp_upgrade/idl"

	"github.com/greenplum-db/gp-common-go-libs/gplog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"os"
)

type AgentServer struct {
	GetDiskUsage  func() (map[string]float64, error)
	commandExecer helpers.CommandExecer
	conf          AgentConfig

	mu      sync.Mutex
	server  *grpc.Server
	lis     net.Listener
	stopped chan struct{}
}

type AgentConfig struct {
	Port     int
	StateDir string
}

func NewAgentServer(execer helpers.CommandExecer, conf AgentConfig) *AgentServer {
	return &AgentServer{
		GetDiskUsage:  diskUsage,
		commandExecer: execer,
		conf:          conf,
		stopped:       make(chan struct{}, 1),
	}
}

func (a *AgentServer) Start() {
	gplog.Error("something")
	createIfNotExists(a.conf.StateDir)
	lis, err := net.Listen("tcp", ":"+strconv.Itoa(a.conf.Port))
	if err != nil {
		gplog.Fatal(err, "failed to listen")
	}

	server := grpc.NewServer()
	a.mu.Lock()
	a.server = server
	a.lis = lis
	a.mu.Unlock()

	pb.RegisterAgentServer(server, a)
	reflection.Register(server)

	err = server.Serve(lis)
	if err != nil {
		gplog.Fatal(err, "failed to serve", err)
	}

	a.stopped <- struct{}{}
}

func (a *AgentServer) Stop() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.server != nil {
		a.server.Stop()
		<-a.stopped
	}
}

func createIfNotExists(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.Mkdir(dir, 0777)
	}
}
