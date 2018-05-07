package services_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"gp_upgrade/hub/services"
	pb "gp_upgrade/idl"
	"gp_upgrade/testutils"

	"google.golang.org/grpc"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gp_upgrade/utils"
)

var _ = Describe("UpgradeShareOids", func() {
	var (
		reader        *testutils.SpyReader
		hub           *services.HubClient
		dir           string
		commandExecer *testutils.FakeCommandExecer
		errChan       chan error
		outChan       chan []byte
	)

	BeforeEach(func() {
		reader = &testutils.SpyReader{
			Hostnames: []string{"hostone", "hosttwo"},
		}

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

		hub = services.NewHub(nil, reader, grpc.DialContext, commandExecer.Exec, &services.HubConfig{
			StateDir: dir,
		})
	})

	AfterEach(func() {
		utils.System = utils.InitializeSystemFunctions()
		os.RemoveAll(dir)
	})

	It("copies files to each host", func() {
		_, err := hub.UpgradeShareOids(nil, &pb.UpgradeShareOidsRequest{})
		Expect(err).ToNot(HaveOccurred())

		hostnames, err := reader.GetHostnames()
		Expect(err).ToNot(HaveOccurred())

		Eventually(commandExecer.GetNumInvocations).Should(Equal(len(hostnames)))

		Expect(commandExecer.Calls()).To(Equal([]string{
			fmt.Sprintf("bash -c rsync -rzpogt %s/pg_upgrade/pg_upgrade_dump_*_oids.sql gpadmin@hostone:%s/pg_upgrade", dir, dir),
			fmt.Sprintf("bash -c rsync -rzpogt %s/pg_upgrade/pg_upgrade_dump_*_oids.sql gpadmin@hosttwo:%s/pg_upgrade", dir, dir),
		}))
	})

	It("copies all files even if rsync fails for a host", func() {
		errChan <- errors.New("failure")

		_, err := hub.UpgradeShareOids(nil, &pb.UpgradeShareOidsRequest{})
		Expect(err).ToNot(HaveOccurred())

		hostnames, err := reader.GetHostnames()
		Expect(err).ToNot(HaveOccurred())

		Eventually(commandExecer.GetNumInvocations).Should(Equal(len(hostnames)))
	})
})
