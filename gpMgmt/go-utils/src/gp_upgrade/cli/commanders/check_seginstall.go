package commanders

import (
	"context"

	pb "gp_upgrade/idl"
)

type SeginstallChecker struct {
	client pb.CliToHubClient
}

func NewSeginstallChecker(client pb.CliToHubClient) SeginstallChecker {
	return SeginstallChecker{
		client: client,
	}
}

func (req SeginstallChecker) Execute() error {
	_, err := req.client.CheckSeginstall(
		context.Background(),
		&pb.CheckSeginstallRequest{},
	)
	return err
}
