package services_test

import (
	"testing"

	"gp_upgrade/utils"

	"github.com/greenplum-db/gp-common-go-libs/testhelper"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCommands(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Hub Services Suite")
}

var _ = BeforeSuite(func() {
	testhelper.SetupTestLogger()
})

var _ = AfterEach(func() {
	utils.System = utils.InitializeSystemFunctions()
})
