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

var _ = Describe("SchemaDeploymentController", func() {
	const timeout = time.Second * 30
	const interval = time.Second * 1

	const templateName = "master-test-template"
	const kqlCfgName = "dev-template-kql"
	const kqlCfgNamespace = "default"
	Context("template with multiple dbs", func() {
		It("should deploy the schmea on all dbs", func() {

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

			spec := kutoschemav1.SchemaDeploymentSpec{

				ApplyTo: kutoschemav1.TargetFilter{
					ClusterUris: []string{"https://sampleadx.westeurope.kusto.windows.net"},
					DB:          "db11",
				},
				Type:           kutoschemav1.DBTypeKusto,
				FailurePolicy:  "abort",
				FailIfDataLoss: true,
				Source: schemav1alpha1.NamespacedName{
					Name:      kqlCfgName,
					Namespace: kqlCfgNamespace,
				},
			}
			key := types.NamespacedName{
				Name:      templateName,
				Namespace: "default",
			}

			toCreate := &kutoschemav1.SchemaDeployment{
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

			By("Creating the versioned deployment successfully")
			time.Sleep(time.Second * 4)
			fetched := &kutoschemav1.SchemaDeployment{}
			Eventually(func() bool {
				err := k8sClient.Get(context.Background(), key, fetched)
				Expect(err).NotTo(HaveOccurred())
				vd := &kutoschemav1.VersionedDeplyment{}
				err = k8sClient.Get(context.Background(), types.NamespacedName(fetched.Status.CurrentVerDeployment), vd)
				return err == nil
			}, timeout, interval).Should(BeTrue())
			fmt.Fprintf(GinkgoWriter, "generated config file: %v\n", fetched.Status.CurrentVerDeployment)
			By("Updating the configMap successfully")
			// 			Expect(k8sClient.Get(context.Background(), types.NamespacedName{
			// 				Name:      kqlCfgName,
			// 				Namespace: kqlCfgNamespace,
			// 			}, cfgToCreate)).Should(Succeed())
			// 			fmt.Fprintf(GinkgoWriter, "ownership config file: %+v\n", cfgToCreate)
			err := k8sClient.Get(context.Background(), types.NamespacedName(fetched.Spec.Source), cfgToCreate)
			Expect(err).NotTo(HaveOccurred())
			cfgToCreate.Data["kql"] = `
			.create-or-alter function  Add(a:real,b:real) {a+b}
			.create-or-alter function  Sub(a:real,b:real) {a-b}
			`
			Expect(k8sClient.Update(context.Background(), cfgToCreate)).Should(Succeed())
			time.Sleep(time.Second * 5)
			fetched = &kutoschemav1.SchemaDeployment{}
			Eventually(func() bool {
				err := k8sClient.Get(context.Background(), key, fetched)
				Expect(err).NotTo(HaveOccurred())
				return fetched.Status.CurrentRevision == 1
			}, timeout, interval).Should(BeTrue())
		})
		It("should deploy the DacPac on all the schmeas ", func() {
			By("Creating the SQL Schema deployment successfully")
			// TODO: create SQL test case
			By("Createing a clustered executer and applying")
			// TODO: define execution verification
		})
	})
})
