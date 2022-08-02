package kustoutils_test

// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
import (
	"fmt"
	"io/ioutil"

	schemav1alpha1 "github.com/microsoft/azure-schema-operator/api/v1alpha1"
	"github.com/microsoft/azure-schema-operator/pkg/kustoutils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
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
		It("Should generate configuration from targets", func() {
			client := &mockKusto{}
			mockClient := &kustoutils.KustoCluster{
				Client: client,
			}
			targets := schemav1alpha1.ClusterTargets{
				DBs: []string{"db1", "db2", "db3"},
			}
			cfgMap := &v1.ConfigMap{
				Data: map[string]string{"kql": "add tables and stuff"},
			}
			failIfDataLoss := false
			exeCfg, err := mockClient.CreateExecConfiguration(targets, cfgMap, failIfDataLoss)
			Expect(err).NotTo(HaveOccurred())
			fmt.Fprintf(GinkgoWriter, "generated config: %v\n", exeCfg)

		})
		It("Should generate configuration from targets", func() {
			client := &mockKusto{}
			mockClient := &kustoutils.KustoCluster{
				Client: client,
			}
			targets := schemav1alpha1.ClusterTargets{
				DBs: []string{"db1", "db2", "db3"},
			}
			cfgMap := &v1.ConfigMap{}
			failIfDataLoss := false
			_, err := mockClient.CreateExecConfiguration(targets, cfgMap, failIfDataLoss)
			Expect(err).To(HaveOccurred())
		})

	})
})
