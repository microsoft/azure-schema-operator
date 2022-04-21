package kustoutils_test

// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
import (
	"fmt"
	"strings"

	"github.com/microsoft/azure-schema-operator/pkg/kustoutils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
)

var _ = Describe("DeltaWrapper", func() {
	Context("when creating configuration", func() {
		cluster := strings.TrimSpace(viper.GetString("schemaop_test_kusto_cluster_name"))
		It("Should generate configuration", func() {
			w := kustoutils.NewDeltaWrapper()
			uri := "https://" + cluster + ".westeurope.kusto.windows.net"
			dbs := []string{"db1", "db2", "db3"}
			kqlFile := "/Users/jocohe/Documents/delta-kusto/dev-state.kql"
			fileName, err := w.CreateExecConfiguration(uri, dbs, kqlFile, true)
			Expect(err).NotTo(HaveOccurred())
			fmt.Fprintf(GinkgoWriter, "generated config file: %s\n", fileName)
			By("Should execute the generated job file", func() {
				err = kustoutils.RunDeltaKusto(fileName)
				Expect(err).NotTo(HaveOccurred())
			})
		})

	})
})
