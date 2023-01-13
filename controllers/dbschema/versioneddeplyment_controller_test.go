package dbschema

// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	kutoschemav1 "github.com/microsoft/azure-schema-operator/apis/dbschema/v1alpha1"
	schemav1alpha1 "github.com/microsoft/azure-schema-operator/apis/dbschema/v1alpha1"
)

var _ = Describe("VersioneddeplymentController", func() {
	const timeout = time.Second * 30
	const interval = time.Second * 10

	const templateName = "versioned-dep-template"
	const kqlCfgName = "dev-versioned-kql"
	const kqlCfgNamespace = "default"
	Context("with new cluster executer CRD", func() {
		It("Should execute the change", func() {
			cfgToCreate := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      kqlCfgName,
					Namespace: kqlCfgNamespace,
				},
				Data: map[string]string{
					"kql": `
.create-or-alter function  Add(a:real,b:real) {a+b}
				`,
				},
			}

			spec := kutoschemav1.VersionedDeplymentSpec{
				ConfigMapName: schemav1alpha1.NamespacedName{
					Name:      kqlCfgName,
					Namespace: kqlCfgNamespace,
				},
				ApplyTo: kutoschemav1.TargetFilter{
					ClusterUris: []string{"https://" + testCluster + ".westeurope.kusto.windows.net"},
					DB:          "db1337",
				},
				Type: schemav1alpha1.DBTypeKusto,
			}
			key := types.NamespacedName{
				Name:      templateName,
				Namespace: "default",
			}

			toCreate := &kutoschemav1.VersionedDeplyment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Spec: spec,
			}
			By("Creating the VersionedDeplyment CRDs successfully")
			Expect(k8sClient.Create(context.Background(), cfgToCreate)).Should(Succeed())
			time.Sleep(time.Second * 3)
			Expect(k8sClient.Create(context.Background(), toCreate)).Should(Succeed())
			time.Sleep(time.Second * 4)
			By("deploying the ClusterExecuter")
			fetched := &kutoschemav1.VersionedDeplyment{}
			Eventually(func() bool {
				err := k8sClient.Get(context.Background(), key, fetched)
				Expect(err).NotTo(HaveOccurred())
				fmt.Fprintf(GinkgoWriter, "versioned deployment status: %+v \n", fetched.Status)
				return len(fetched.Status.Executers) == 1
			}, timeout, interval).Should(BeTrue())
			// By("waiting for the ClusterExecuter")
			// ce := &kutoschemav1.ClusterExecuter{}
			// Eventually(func() bool {
			// 	err := k8sClient.Get(context.Background(), types.NamespacedName(fetched.Status.Executers[0]), ce)
			// 	Expect(err).NotTo(HaveOccurred())
			// 	fmt.Fprintf(GinkgoWriter, "versioned deployment - fetched ClusterExecuter: %+v \n", ce)
			// 	return ce.Status.Executed
			// }, timeout, interval).Should(BeTrue())
			// fmt.Fprintf(GinkgoWriter, "versioned deployment - fetched ClusterExecuter status: %+v \n", ce.Status)
			// By("checking the updates versioned deployment status")
			// err := k8sClient.Get(context.Background(), key, fetched)
			// Expect(err).NotTo(HaveOccurred())
			// fmt.Fprintf(GinkgoWriter, "versioned deployment status: %+v \n", fetched.Status)
			// Expect(fetched.Status.Failed).To(Equal(int32(0)))
			// Expect(fetched.Status.Succeeded).To(Equal(int32(1)))
		})
	})
})
