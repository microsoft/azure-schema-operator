package utils_test

import (
	"testing"

	"os"

	"github.com/microsoft/azure-schema-operator/pkg/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestUtils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Utils Suite")
}

var _ = Describe("Utils", func() {
	Context("with a temp file ", func() {
		It("Should remove the file", func() {
			f, err := os.CreateTemp("/tmp", "test-*.yaml")
			Expect(err).NotTo(HaveOccurred())
			err = utils.CleanupFile(f.Name())
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
