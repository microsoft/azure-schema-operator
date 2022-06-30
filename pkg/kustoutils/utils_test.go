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
	"github.com/Azure/go-autorest/autorest"
	schemav1alpha1 "github.com/microsoft/azure-schema-operator/api/v1alpha1"
	"github.com/microsoft/azure-schema-operator/pkg/kustoutils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type mockKusto struct {
}

func (m *mockKusto) Close() error {
	return nil
}

func (m *mockKusto) Auth() kusto.Authorization {
	return kusto.Authorization{Authorizer: autorest.NewBasicAuthorizer("", "")}
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
	ri.Mock(mr)
	return ri, nil
}

func (m *mockKusto) HttpClient() *http.Client {
	return &http.Client{}
}

var _ = Describe("Utils", func() {
	// add utils tests
	Context("fake until we make it", func() {
		var mockClient *kustoutils.KustoCluster
		It("should mock client", func() {
			client := &mockKusto{}
			mockClient = &kustoutils.KustoCluster{
				Client: client,
			}
			expression := ".*"
			dbs, err := mockClient.ListDatabases(expression)
			Expect(err).NotTo(HaveOccurred())
			fmt.Fprintf(GinkgoWriter, "List of DBs in Mock Cluster: %+v \n", dbs)
			Expect(len(dbs)).To(Equal(2))
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
