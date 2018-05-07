package services_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	pb "gp_upgrade/idl"
	"gp_upgrade/testutils"
	"gp_upgrade/utils"

	"gp_upgrade/agent/services"

	"github.com/greenplum-db/gp-common-go-libs/testhelper"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CommandListener", func() {
	var (
		agent         *services.AgentServer
		dir           string
		commandExecer *testutils.FakeCommandExecer
		oidFile       string
		errChan       chan error
		outChan       chan []byte
	)

	BeforeEach(func() {
		testhelper.SetupTestLogger()

		var err error
		dir, err = ioutil.TempDir("", "")
		Expect(err).ToNot(HaveOccurred())

		errChan = make(chan error, 2)
		outChan = make(chan []byte, 2)
		commandExecer = &testutils.FakeCommandExecer{}
		commandExecer.SetOutput(&testutils.FakeCommand{
			Err: errChan,
			Out: outChan,
		})

		agentConfig := services.AgentConfig{StateDir: dir}
		agent = services.NewAgentServer(commandExecer.Exec, agentConfig)

		err = os.MkdirAll(filepath.Join(dir, "pg_upgrade"), 0700)
		Expect(err).ToNot(HaveOccurred())

		oidFile = filepath.Join(dir, "pg_upgrade", "pg_upgrade_dump_seg1_oids.sql")
		f, err := os.Create(oidFile)
		Expect(err).ToNot(HaveOccurred())
		f.Close()
	})

	AfterEach(func() {
		utils.System = utils.InitializeSystemFunctions()
	})

	It("successfully runs pg_upgrade", func() {
		cmd := &testutils.FakeCommand{}
		commandExecer.SetOutput(cmd)

		_, err := agent.UpgradeConvertPrimarySegments(nil, &pb.UpgradeConvertPrimarySegmentsRequest{
			OldBinDir: "/old/bin",
			NewBinDir: "/new/bin",
			DataDirPairs: []*pb.DataDirPair{
				{OldDataDir: "old/datadir1", NewDataDir: "new/datadir1", Content: 0, OldPort: 1, NewPort: 11},
				{OldDataDir: "old/datadir2", NewDataDir: "new/datadir2", Content: 1, OldPort: 2, NewPort: 22},
			},
		})
		Expect(err).ToNot(HaveOccurred())

		Expect(commandExecer.GetNumInvocations()).To(Equal(4))

		Expect(commandExecer.Calls()).To(ContainElement(fmt.Sprintf("cp %s %s/pg_upgrade/seg-0", oidFile, dir)))
		Expect(commandExecer.Calls()).To(ContainElement(fmt.Sprintf("bash -c cd %s/pg_upgrade/seg-0 && nohup /new/bin/pg_upgrade --old-bindir=/old/bin --old-datadir=old/datadir1 --new-bindir=/new/bin --new-datadir=new/datadir1 --old-port=1 --new-port=11 --progress", dir)))
		Expect(commandExecer.Calls()).To(ContainElement(fmt.Sprintf("cp %s %s/pg_upgrade/seg-1", oidFile, dir)))
		Expect(commandExecer.Calls()).To(ContainElement(fmt.Sprintf("bash -c cd %s/pg_upgrade/seg-1 && nohup /new/bin/pg_upgrade --old-bindir=/old/bin --old-datadir=old/datadir2 --new-bindir=/new/bin --new-datadir=new/datadir2 --old-port=2 --new-port=22 --progress", dir)))
	})

	It("returns an an error if the oid files glob fails", func() {
		utils.System.FilePathGlob = func(pattern string) ([]string, error) {
			return []string{}, errors.New("failed to find files")
		}

		_, err := agent.UpgradeConvertPrimarySegments(nil, &pb.UpgradeConvertPrimarySegmentsRequest{})
		Expect(err).To(HaveOccurred())
	})

	It("returns an an error if no oid files are found", func() {
		err := os.Remove(oidFile)
		Expect(err).ToNot(HaveOccurred())

		_, err = agent.UpgradeConvertPrimarySegments(nil, &pb.UpgradeConvertPrimarySegmentsRequest{})
		Expect(err).To(HaveOccurred())
	})

	It("returns an error if the pg_upgrade/segmentDir cannot be made", func() {
		utils.System.MkdirAll = func(path string, perm os.FileMode) error {
			return errors.New("failed to create segment directory")
		}

		_, err := agent.UpgradeConvertPrimarySegments(nil, &pb.UpgradeConvertPrimarySegmentsRequest{
			OldBinDir: "/old/bin",
			NewBinDir: "/new/bin",
			DataDirPairs: []*pb.DataDirPair{
				{OldDataDir: "old/datadir1", NewDataDir: "new/datadir1", Content: 0, OldPort: 1, NewPort: 11},
			},
		})
		Expect(err).To(HaveOccurred())
	})

	It("returns an error if the oid files fail to copy into the segment directory", func() {
		errChan <- errors.New("Failed to copy oid file into segment directory")

		_, err := agent.UpgradeConvertPrimarySegments(nil, &pb.UpgradeConvertPrimarySegmentsRequest{
			OldBinDir: "/old/bin",
			NewBinDir: "/new/bin",
			DataDirPairs: []*pb.DataDirPair{
				{OldDataDir: "old/datadir1", NewDataDir: "new/datadir1", Content: 0, OldPort: 1, NewPort: 11},
			},
		})
		Expect(err).To(HaveOccurred())
	})

	It("returns an error if starting pg_upgrade fails", func() {
		errChan <- nil
		errChan <- errors.New("convert primary on agent failed")

		_, err := agent.UpgradeConvertPrimarySegments(nil, &pb.UpgradeConvertPrimarySegmentsRequest{
			OldBinDir: "/old/bin",
			NewBinDir: "/new/bin",
			DataDirPairs: []*pb.DataDirPair{
				{OldDataDir: "old/datadir1", NewDataDir: "new/datadir1", Content: 0, OldPort: 1, NewPort: 11},
			},
		})
		Expect(err).To(HaveOccurred())
	})
})
