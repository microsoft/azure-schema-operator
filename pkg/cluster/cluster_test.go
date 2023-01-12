package cluster_test

// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
import (
	schemav1alpha1 "github.com/microsoft/azure-schema-operator/apis/dbschema/v1alpha1"
	"github.com/microsoft/azure-schema-operator/pkg/cluster"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cluster", func() {
	Context("Testing the different helper utils", func() {
		a := schemav1alpha1.ClusterTargets{
			DBs:     []string{"tenant_1", "tenant_2"},
			Schemas: []string{},
		}
		b := schemav1alpha1.ClusterTargets{
			DBs:     []string{"tenant_2", "tenant_3"},
			Schemas: []string{},
		}

		c := cluster.Difference(a, b)
		Expect(len(c.DBs)).To(Equal(1))
		u := cluster.Union(a, b)
		Expect(len(u.DBs)).To(Equal(3))

	})
})
