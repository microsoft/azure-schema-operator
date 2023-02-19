package kustoutils_test

import (
	"context"

	"github.com/microsoft/azure-schema-operator/pkg/kustoutils"
	"github.com/microsoft/azure-schema-operator/pkg/kustoutils/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Azure/azure-kusto-go/kusto"
)

var _ = Describe("Policymanager", Label("Policymanager"), Label("live"), func() {
	var (
		client    *kusto.Client
		database  string
		tableName string
		err       error
	)

	BeforeEach(func() {
		GinkgoWriter.Println("connecting to cluster: ", testCluster)
		kcsb := kusto.NewConnectionStringBuilder(testCluster).WithDefaultAzureCredential()
		client, err = kusto.New(kcsb)
		Expect(err).NotTo(HaveOccurred())
		database = "test"
		tableName = "test"
	})
	It("should get a policy", func() {
		ctx := context.Background()
		retentionPolicy, err := kustoutils.GetTableRetentionPolicy(ctx, client, database, tableName)
		Expect(err).NotTo(HaveOccurred())
		GinkgoWriter.Println("Test table retuention policy:", retentionPolicy)

	})
	It("should get the db policy for an empty table ", func() {
		ctx := context.Background()
		retentionPolicy, err := kustoutils.GetTableRetentionPolicy(ctx, client, database, "")
		Expect(err).NotTo(HaveOccurred())
		GinkgoWriter.Println("Test database retuention policy:", retentionPolicy)

	})
	It("should fail to get a policy for non existing table", func() {
		ctx := context.Background()
		retentionPolicy, err := kustoutils.GetTableRetentionPolicy(ctx, client, database, "doesnotexist")
		Expect(err).To(HaveOccurred())
		Expect(retentionPolicy).To(BeNil())
	})
	It("should set a policy", func() {
		ctx := context.Background()
		newPolicy := &types.RetentionPolicy{SoftDeletePeriod: "12.00:00:00", Recoverability: "Enabled"}
		retentionPolicy, err := kustoutils.SetTableRetentionPolicy(ctx, client, database, tableName, newPolicy)
		Expect(err).NotTo(HaveOccurred())
		GinkgoWriter.Println("Test table retuention policy:", retentionPolicy)

	})
	It("should get a caching policy", Label("cachingpolicy"), func() {
		ctx := context.Background()
		policy := &types.CachingPolicy{}
		err := kustoutils.GetTablePolicy(ctx, client, database, tableName, policy)
		Expect(err).NotTo(HaveOccurred())
		GinkgoWriter.Println("Test table policy:", policy)

	})
	It("should set a caching policy", Label("cachingpolicy"), func() {
		ctx := context.Background()
		newPolicy := "7d"
		cachingPolicy, err := kustoutils.SetTableCachingPolicy(ctx, client, database, tableName, newPolicy)
		Expect(err).NotTo(HaveOccurred())
		GinkgoWriter.Println("Test table retuention policy:", cachingPolicy)

	})
})
