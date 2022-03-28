package controllers

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

	kutoschemav1 "github.com/microsoft/azure-schema-operator/api/v1alpha1"
	schemav1alpha1 "github.com/microsoft/azure-schema-operator/api/v1alpha1"
)

var _ = Describe("VersioneddeplymentController", func() {
	// const timeout = time.Second * 30
	// const interval = time.Second * 3

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
					ClusterUris: []string{"https://sampleadx.westeurope.kusto.windows.net"},
					DB:          "db11",
				},
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
			By("Creating the template successfully")
			Expect(k8sClient.Create(context.Background(), cfgToCreate)).Should(Succeed())
			time.Sleep(time.Second * 3)
			Expect(k8sClient.Create(context.Background(), toCreate)).Should(Succeed())
			time.Sleep(time.Second * 4)
			By("deploying the ClusterExecuter")
			fetched := &kutoschemav1.VersionedDeplyment{}
			fmt.Fprintf(GinkgoWriter, "versioned deployment status: %+v \n", fetched.Status)
			// Eventually(func() bool {
			// 	k8sClient.Get(context.Background(), key, fetched)
			// 	fmt.Fprintf(GinkgoWriter, "versioned deployment status: %+v \n", fetched.Status)
			// 	ce := &kutoschemav1.ClusterExecuter{}
			// 	err := k8sClient.Get(context.Background(), types.NamespacedName(fetched.Status.Executers[0]), ce)
			// 	return err == nil
			// }, timeout, interval).Should(BeTrue())
			// Expect(fetched.Status.Failed).To(Equal(int32(0)))
			// Expect(fetched.Status.Succeeded).To(Equal(int32(1)))
		})
	})
})
