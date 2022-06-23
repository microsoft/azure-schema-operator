package kustoutils_test

// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
import (
	"fmt"

	"github.com/Azure/azure-kusto-go/kusto"
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
})
