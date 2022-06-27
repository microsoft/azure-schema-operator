package kustoutils_test

// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
import (
	"fmt"

	"github.com/Azure/azure-kusto-go/kusto"
	schemav1alpha1 "github.com/microsoft/azure-schema-operator/api/v1alpha1"
	"github.com/microsoft/azure-schema-operator/pkg/kustoutils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Utils", func() {
	// add utils tests
	Context("fake until we make it", func() {
		var mockClient *kustoutils.KustoCluster
		It("should mock client", func() {
			client := kusto.NewMockClient()
			mockClient = &kustoutils.KustoCluster{
				Client: client,
			}
			expression := ".*"
			dbs, err := mockClient.ListDatabases(expression)
			Expect(err).NotTo(HaveOccurred())
			fmt.Fprintf(GinkgoWriter, "List of DBs in Mock Cluster: %+v \n", dbs)
			Expect(len(dbs)).To(Equal(0))
			//TODO - mock some dbs into this so we can actually test...

		})
	})
	if liveTest {
		Context("when testing kusto with a live server", func() {
			ClusterUri := "https://" + testCluster + ".westeurope.kusto.windows.net"
			cluster := kustoutils.NewKustoCluster(ClusterUri)
			filter := schemav1alpha1.TargetFilter{
				DB: "db1948",
			}
			It("Should acquire requested targets and prepare for execution", func() {
				clusterTargets, err := cluster.AquireTargets(filter)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(clusterTargets.DBs)).To(Equal(1))
			})
		})
	}
})
