package kustoutils_test

// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/microsoft/azure-schema-operator/pkg/kustoutils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
)

const tstCfgContent = `
sendErrorOptIn: false
failIfDataLoss: true
jobs:
  push-db1-to-prod:
    current:
      adx:
        clusterUri:  https://testcluster.westeurope.kusto.windows.net 
        database: db1
    target:
      scripts:
        - filePath: /path/to/schema.kql 
    action:
      # filePath: prod-update.kql
      pushToCurrent: true
  push-db2-to-prod:
    current:
      adx:
        clusterUri:  https://testcluster.westeurope.kusto.windows.net 
        database: db2
    target:
      scripts:
        - filePath: /path/to/schema.kql 
    action:
      # filePath: prod-update.kql
      pushToCurrent: true
  push-db3-to-prod:
    current:
      adx:
        clusterUri:  https://testcluster.westeurope.kusto.windows.net 
        database: db3
    target:
      scripts:
        - filePath: /path/to/schema.kql 
    action:
      # filePath: prod-update.kql
      pushToCurrent: true
tokenProvider:
  login:
    tenantId: to-be-overridden
    clientId: to-be-overridden
    secret: to-be-overridden
`

var _ = Describe("DeltaWrapper", func() {
	Context("when creating configuration", func() {
		cluster := strings.TrimSpace(viper.GetString("schemaop_test_kusto_cluster_name"))
		It("Should generate configuration", func() {
			w := kustoutils.NewDeltaWrapper()
			uri := "https://testcluster.westeurope.kusto.windows.net"
			dbs := []string{"db1", "db2", "db3"}
			kqlFile := "/path/to/schema.kql"

			fileName, err := w.CreateExecConfiguration(uri, dbs, kqlFile, true)
			Expect(err).NotTo(HaveOccurred())
			Expect(fileName).To(BeARegularFile())
			fmt.Fprintf(GinkgoWriter, "generated config file: %s\n", fileName)
			b, err := ioutil.ReadFile(fileName) // just pass the file name
			Expect(err).NotTo(HaveOccurred())
			genCfgStr := string(b) // convert content to a 'string'
			Expect(genCfgStr).To(Equal(tstCfgContent))

		})

	})
})
