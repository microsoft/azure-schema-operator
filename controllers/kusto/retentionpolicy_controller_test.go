package kusto

// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	kustov1alpha1 "github.com/microsoft/azure-schema-operator/apis/kusto/v1alpha1"
	kustotypes "github.com/microsoft/azure-schema-operator/pkg/kustoutils/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("RetentionpolicyController", func() {
	const timeout = time.Second * 30
	const interval = time.Second * 10

	const policyNamw = "test-table-policy"
	Context("When creating a retention policy", func() {
		It("should create a retention policy", func() {

			spec := kustov1alpha1.RetentionPolicySpec{
				PolicySpec: kustov1alpha1.PolicySpec{
					ClusterUris: []string{testCluster},
					DB:          "test",
					Table:       "test",
				},
				RetentionPolicy: kustotypes.RetentionPolicy{
					SoftDeletePeriod: "15.00:00:00",
					Recoverability:   "Enabled",
				},
			}

			key := types.NamespacedName{
				Name:      policyNamw,
				Namespace: "default",
			}

			toCreate := &kustov1alpha1.RetentionPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Spec: spec,
			}
			By("Creating the RetentionPolicy CRDs successfully")
			Expect(k8sClient.Create(context.Background(), toCreate)).Should(Succeed())
			By("applying the RetentionPolicy")
			fetched := &kustov1alpha1.RetentionPolicy{}
			Eventually(func() bool {
				err := k8sClient.Get(context.Background(), key, fetched)
				Expect(err).NotTo(HaveOccurred())
				fmt.Fprintf(GinkgoWriter, "Retention Policy status: %+v \n", fetched.Status)
				return len(fetched.Status.ClustersDone) == 1
			}, timeout, interval).Should(BeTrue())

		})
	})

})
