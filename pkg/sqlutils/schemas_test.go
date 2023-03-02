package sqlutils_test

// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	schemav1alpha1 "github.com/microsoft/azure-schema-operator/apis/dbschema/v1alpha1"
	"github.com/microsoft/azure-schema-operator/pkg/sqlutils"
	v1 "k8s.io/api/core/v1"
)

var _ = Describe("Schemas", func() {
	Context("When Processing a cluster", func() {
		cluster := sqlutils.NewSQLCluster("fakecluster.database.windows.net", nil, nil)
		filter := schemav1alpha1.TargetFilter{
			DB: "DB1",
		}
		cfgMap := &v1.ConfigMap{
			Data: map[string]string{
				"templateName":      "tenant_",
				"sqlpackageOptions": "/p:option1:value /p:option1:value2",
			},
			BinaryData: map[string][]byte{
				"dacpac": []byte("not really a dacpac"),
			},
		}
		It("Should acquire requested targets and prepare for execution", func() {
			clusterTargets, err := cluster.AquireTargets(filter)
			Expect(err).NotTo(HaveOccurred())
			executionConfiguration, err := cluster.CreateExecConfiguration(clusterTargets, cfgMap, true)
			Expect(err).NotTo(HaveOccurred())
			Expect(executionConfiguration.DacPac).To(ContainSubstring("/tmp/"))
			Expect(executionConfiguration.Properties["sqlpackageOptions"]).To(ContainSubstring("/p:BlockOnPossibleDataLoss=true"))
		})

	})
	if liveTest {
		Context("when testing sqlpackage with a live server", Label("live"), func() {
			cluster := sqlutils.NewSQLCluster(testCluster+".database.windows.net", nil, nil)
			filter := schemav1alpha1.TargetFilter{
				DB:     "DB1",
				Schema: "db1111",
			}
			It("Should acquire requested targets and prepare for execution", func() {
				clusterTargets, err := cluster.AquireTargets(filter)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(clusterTargets.Schemas)).To(Equal(1))
			})
		})
	}

})
