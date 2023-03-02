package kustoutils_test

// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
import (
	"context"
	"fmt"
	"net/http"

	"github.com/Azure/azure-kusto-go/kusto"
	"github.com/Azure/azure-kusto-go/kusto/data/table"
	"github.com/Azure/azure-kusto-go/kusto/data/types"
	"github.com/Azure/azure-kusto-go/kusto/data/value"
	schemav1alpha1 "github.com/microsoft/azure-schema-operator/apis/dbschema/v1alpha1"
	"github.com/microsoft/azure-schema-operator/pkg/kustoutils"
	"github.com/microsoft/azure-schema-operator/pkg/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type mockKusto struct {
}

func (m *mockKusto) Close() error {
	return nil
}

func (m *mockKusto) Auth() kusto.Authorization {
	// return kusto.Authorization{Authorizer: autorest.NewBasicAuthorizer("", "")}
	return kusto.Authorization{}
}

func (m *mockKusto) Endpoint() string {
	return "https://mock.eastus.kusto.windows.net"
}

func (m *mockKusto) Query(ctx context.Context, db string, query kusto.Stmt, options ...kusto.QueryOption) (*kusto.RowIterator, error) {
	panic("not implemented") // TODO: Implement
}

func (m *mockKusto) Mgmt(ctx context.Context, db string, query kusto.Stmt, options ...kusto.MgmtOption) (*kusto.RowIterator, error) {
	columns := table.Columns{
		{Name: "DatabaseName", Type: types.String},
	}
	rows := []value.Values{
		{value.String{Valid: true, Value: "tenant_1"}},
		{value.String{Valid: true, Value: "tenant_2"}},
	}
	mr, err := kusto.NewMockRows(columns)
	if err != nil {
		panic(err) // This panic and all others are setup errors, not test errors
	}
	for _, row := range rows {
		if err := mr.Row(row); err != nil {
			panic(err)
		}
	}
	ri := &kusto.RowIterator{}
	err = ri.Mock(mr)
	return ri, err
}

func (m *mockKusto) HttpClient() *http.Client {
	return &http.Client{}
}

var _ = Describe("Utils", func() {
	// add utils tests
	Context("Filter results from mock client", func() {
		var mockClient *kustoutils.KustoCluster
		It("should return all dbs", func() {
			client := &mockKusto{}
			mockClient = &kustoutils.KustoCluster{
				Client: client,
			}
			expression := ".*"
			dbs, err := mockClient.ListDatabases(expression)
			Expect(err).NotTo(HaveOccurred())
			fmt.Fprintf(GinkgoWriter, "List of DBs in Mock Cluster: %+v \n", dbs)
			Expect(len(dbs)).To(Equal(2))
		})
		It("should return only filtered results", func() {
			client := &mockKusto{}
			mockClient = &kustoutils.KustoCluster{
				Client: client,
			}
			expression := "tenant_1"
			dbs, err := mockClient.ListDatabases(expression)
			Expect(err).NotTo(HaveOccurred())
			fmt.Fprintf(GinkgoWriter, "List of DBs in Mock Cluster: %+v \n", dbs)
			Expect(len(dbs)).To(Equal(1))
		})
		It("should filter everything", func() {
			client := &mockKusto{}
			mockClient = &kustoutils.KustoCluster{
				Client: client,
			}
			expression := "db_1"
			dbs, err := mockClient.ListDatabases(expression)
			Expect(err).NotTo(HaveOccurred())
			fmt.Fprintf(GinkgoWriter, "List of DBs in Mock Cluster: %+v \n", dbs)
			Expect(len(dbs)).To(Equal(0))
		})
		It("should AquireTargets filtered results", func() {
			client := &mockKusto{}
			mockClient = &kustoutils.KustoCluster{
				Client: client,
			}
			filter := schemav1alpha1.TargetFilter{
				DB: "tenant_1",
			}
			targets, err := mockClient.AquireTargets(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(len(targets.DBs)).To(Equal(1))
		})
		It("should AquireTargets all dbs without filter", func() {
			client := &mockKusto{}
			mockClient = &kustoutils.KustoCluster{
				Client: client,
			}
			filter := schemav1alpha1.TargetFilter{}
			targets, err := mockClient.AquireTargets(filter)
			Expect(err).NotTo(HaveOccurred())

			Expect(len(targets.DBs)).To(Equal(2))
		})

	})
	if liveTest {
		Context("when testing kusto with a live server", func() {
			ClusterUri := testCluster
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
	Context("when testing fake schema data", func() {
		It("should store the schema in a file", func() {
			schema := "fake schema"
			fileName, err := kustoutils.StoreKQLSchemaToFile(schema)
			Expect(err).NotTo(HaveOccurred())
			Expect(fileName).NotTo(BeNil())
			fmt.Fprintf(GinkgoWriter, "generated config file: %s\n", fileName)
			err = utils.CleanupFile(fileName)
			Expect(err).NotTo(HaveOccurred())
		})
	})
	Context("simple test for utile", Label("utils"), func() {
		It("should get the kusto cluster name from url", func() {
			uri := "https://testcluster.westeurope.kusto.windows.net"
			kustoCluster := kustoutils.ClusterNameFromURI(uri)
			Expect(kustoCluster).To(Equal("testcluster"))
		})
		It("should convert the time string to the expected format", func() {
			timeString := "1.02:00:00"
			expectedTime := "26h"
			convertedTime := kustoutils.ConvertTimeFormat(timeString)
			Expect(convertedTime).To(Equal(expectedTime))
			timeString = "7.0:00:00"
			expectedTime = "7d"
			convertedTime = kustoutils.ConvertTimeFormat(timeString)
			Expect(convertedTime).To(Equal(expectedTime))

		})
	})
})
