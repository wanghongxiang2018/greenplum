package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"

	"gp_upgrade/helpers"
	"gp_upgrade/hub/cluster"
	"gp_upgrade/hub/configutils"
	"gp_upgrade/hub/services"

	"github.com/greenplum-db/gp-common-go-libs/gplog"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// This directory to have the implementation code for the gRPC server to serve
// Minimal CLI command parsing to embrace that booting this binary to run the hub might have some flags like a log dir

func main() {
	var logdir string
	var RootCmd = &cobra.Command{
		Use:   "gp_upgrade_hub [--log-directory path]",
		Short: "Start the gp_upgrade_hub (blocks)",
		Long:  `Start the gp_upgrade_hub (blocks)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			gplog.InitializeLogging("gp_upgrade_hub", logdir)
			debug.SetTraceback("all")

			conf := &services.HubConfig{
				CliToHubPort:   7527,
				HubToAgentPort: 6416,
				StateDir:       filepath.Join(os.Getenv("HOME"), ".gp_upgrade"),
				LogDir:         logdir,
			}
			reader := configutils.NewReader()

			commandExecer := func(command string, vars ...string) helpers.Command {
				return exec.Command(command, vars...)
			}
			hub := services.NewHub(&cluster.Pair{}, &reader, grpc.DialContext, commandExecer, conf)
			hub.Start()

			hub.Stop()

			return nil
		},
	}

	RootCmd.PersistentFlags().StringVar(&logdir, "log-directory", "", "gp_upgrade_hub log directory")

	if err := RootCmd.Execute(); err != nil {
		gplog.Error(err.Error())
		os.Exit(1)
	}
}
