package upgradestatus

import (
	pb "gp_upgrade/idl"
	"gp_upgrade/utils"
	"path/filepath"

	"github.com/pkg/errors"
)

type StateCheck struct {
	path string
	step pb.UpgradeSteps
}

func NewStateCheck(path string, step pb.UpgradeSteps) StateCheck {
	return StateCheck{
		path: path,
		step: step,
	}
}

func (c StateCheck) GetStatus() (*pb.UpgradeStepStatus, error) {
	_, err := utils.System.Stat(c.path)
	if err != nil {
		return &pb.UpgradeStepStatus{Step: c.step, Status: pb.StepStatus_PENDING}, nil
	}
	files, err := utils.System.FilePathGlob(filepath.Join(c.path, "*"))
	if len(files) > 1 {
		return nil, errors.New("got more files than expected")
	} else if len(files) == 1 {
		switch files[0] {
		case filepath.Join(c.path, "failed"):
			return &pb.UpgradeStepStatus{Step: c.step, Status: pb.StepStatus_FAILED}, nil
		case filepath.Join(c.path, "completed"):
			return &pb.UpgradeStepStatus{Step: c.step, Status: pb.StepStatus_COMPLETE}, nil
		case filepath.Join(c.path, "in.progress"):
			return &pb.UpgradeStepStatus{Step: c.step, Status: pb.StepStatus_RUNNING}, nil
		}
	}

	return &pb.UpgradeStepStatus{Step: c.step, Status: pb.StepStatus_PENDING}, nil
}
