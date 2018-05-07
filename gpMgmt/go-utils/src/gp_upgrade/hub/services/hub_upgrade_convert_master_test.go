package services_test

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"gp_upgrade/hub/services"
	pb "gp_upgrade/idl"
	"gp_upgrade/testutils"
	"gp_upgrade/utils"

	"google.golang.org/grpc"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ConvertMasterHub", func() {
	var (
		dir           string
		commandExecer *testutils.FakeCommandExecer
		hub           *services.HubClient
		outChan       chan []byte
		errChan       chan error
		portChan      chan int
	)

	BeforeEach(func() {
		portChan = make(chan int, 2)
		reader := &testutils.SpyReader{
			Port: portChan,
		}

		var err error
		dir, err = ioutil.TempDir("", "")
		Expect(err).ToNot(HaveOccurred())
		conf := &services.HubConfig{
			StateDir: dir,
		}

		errChan = make(chan error, 2)
		outChan = make(chan []byte, 2)
		commandExecer = &testutils.FakeCommandExecer{}
		commandExecer.SetOutput(&testutils.FakeCommand{
			Err: errChan,
			Out: outChan,
		})

		hub = services.NewHub(nil, reader, grpc.DialContext, commandExecer.Exec, conf)
	})

	AfterEach(func() {
		utils.System = utils.InitializeSystemFunctions()
		os.RemoveAll(dir)
	})

	It("returns with no error when convert master runs successfully", func() {
		portChan <- 5432
		portChan <- 6432

		_, err := hub.UpgradeConvertMaster(nil, &pb.UpgradeConvertMasterRequest{
			OldBinDir:  "/old/path/bin",
			OldDataDir: "old/data/dir",
			NewBinDir:  "/new/path/bin",
			NewDataDir: "new/data/dir",
		})
		Expect(err).ToNot(HaveOccurred())

		pgupgrade_dir := filepath.Join(dir, "pg_upgrade")
		Expect(commandExecer.Command()).To(Equal("bash"))
		Expect(commandExecer.Args()).To(Equal([]string{
			"-c",
			"unset PGHOST; unset PGPORT; cd " + pgupgrade_dir +
				` && nohup /new/path/bin/pg_upgrade --old-bindir=/old/path/bin ` +
				`--old-datadir=old/data/dir --new-bindir=/new/path/bin ` +
				`--new-datadir=new/data/dir --old-port=5432 --new-port=6432 --dispatcher-mode --progress`,
		}))
	})

	It("returns an error when one master port cannot be found", func() {
		portChan <- 5432

		_, err := hub.UpgradeConvertMaster(nil, &pb.UpgradeConvertMasterRequest{
			OldBinDir:  "/old/path/bin",
			OldDataDir: "old/data/dir",
			NewBinDir:  "/new/path/bin",
			NewDataDir: "new/data/dir",
		})
		Expect(err).To(HaveOccurred())
	})

	It("returns an error when convert master fails", func() {
		errChan <- errors.New("upgrade failed")

		_, err := hub.UpgradeConvertMaster(nil, &pb.UpgradeConvertMasterRequest{
			OldBinDir:  "/old/path/bin",
			OldDataDir: "old/data/dir",
			NewBinDir:  "/new/path/bin",
			NewDataDir: "new/data/dir",
		})
		Expect(err).To(HaveOccurred())
	})

	It("returns an error if the upgrade directory cannot be created", func() {
		utils.System.MkdirAll = func(path string, perm os.FileMode) error {
			return errors.New("failed to create directory")
		}

		_, err := hub.UpgradeConvertMaster(nil, &pb.UpgradeConvertMasterRequest{
			OldBinDir:  "/old/path/bin",
			OldDataDir: "old/data/dir",
			NewBinDir:  "/new/path/bin",
			NewDataDir: "new/data/dir",
		})
		Expect(err).To(HaveOccurred())
	})

	It("returns an error if the upgrade file cannot be created", func() {
		utils.System.OpenFile = func(name string, flag int, perm os.FileMode) (*os.File, error) {
			return nil, errors.New("failed to open file")
		}

		_, err := hub.UpgradeConvertMaster(nil, &pb.UpgradeConvertMasterRequest{
			OldBinDir:  "/old/path/bin",
			OldDataDir: "old/data/dir",
			NewBinDir:  "/new/path/bin",
			NewDataDir: "new/data/dir",
		})
		Expect(err).To(HaveOccurred())
	})
})
