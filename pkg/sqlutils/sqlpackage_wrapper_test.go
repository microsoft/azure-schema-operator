package sqlutils_test

// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
import (
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"

	"github.com/microsoft/azure-schema-operator/pkg/sqlutils"
)

var _ = Describe("SqlpackageWrapper", func() {
	if liveTest {
		Context("when running a common Dacpac", func() {
			dacpacURL := strings.TrimSpace(viper.GetString("schemaop_test_sqlserver_dacpac"))
			clusterUri := testCluster + ".database.windows.net"
			dbName := "db2"

			if dacpacURL != "" {
				dacpac, err := sqlutils.DownloadDacPacFromURL(dacpacURL)
				Expect(err).NotTo(HaveOccurred())

				It("should execute the original dacpac without errors", func() {
					err := sqlutils.RunDacPac(dacpac, clusterUri, dbName, "")
					Expect(err).To(Not(HaveOccurred()))
				})
				It("Should modify the dacpac and run it", func() {
					success, err := sqlutils.TargetDacpacExecution(clusterUri, dbName, "", dacpac, "TestTenant", "schema1")
					Expect(err).To(Not(HaveOccurred()))
					Expect(success).To(BeTrue())
				})
			}

		})
	}
})
