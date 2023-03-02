package dbschema

// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	kutoschemav1 "github.com/microsoft/azure-schema-operator/apis/dbschema/v1alpha1"
	schemav1alpha1 "github.com/microsoft/azure-schema-operator/apis/dbschema/v1alpha1"
)

var _ = Describe("ClusterexecuterController", func() {
	// const timeout = time.Second * 30
	// const interval = time.Second * 10

	const templateName = "cluster-exec-test"
	const kqlCfgName = "dev-exec-kql"
	const kqlCfgNamespace = "default"
	Context("with new cluster executer CRD", func() {

		It("Should execute the change", func() {
			// re-write this test to check the cluster executer flows
			By("Configuring the CRDs")
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

			spec := kutoschemav1.ClusterExecuterSpec{
				ClusterUri:     testCluster,
				Type:           schemav1alpha1.DBTypeKusto,
				Revision:       1,
				FailIfDataLoss: true,
				ConfigMapName: schemav1alpha1.NamespacedName{
					Name:      kqlCfgName,
					Namespace: kqlCfgNamespace,
				},
				ApplyTo: kutoschemav1.TargetFilter{
					ClusterUris: []string{testCluster},
					DB:          "db1948",
				},
			}
			key := types.NamespacedName{
				Name:      templateName,
				Namespace: "default",
			}

			toCreate := &kutoschemav1.ClusterExecuter{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Spec: spec,
			}
			By("Creating the CRDs successfully")
			Expect(k8sClient.Create(context.Background(), cfgToCreate)).Should(Succeed())
			time.Sleep(time.Second * 3)
			Expect(k8sClient.Create(context.Background(), toCreate)).Should(Succeed())
			// time.Sleep(time.Second * 4)
			// By("waiting for execution to complete")
			// ce := &kutoschemav1.ClusterExecuter{}
			// Eventually(func() bool {
			// 	k8sClient.Get(context.Background(), key, ce)
			// 	fmt.Fprintf(GinkgoWriter, "executer deployment status: %+v \n", ce)
			// 	return ce.Status.Executed
			// }, timeout, interval).Should(BeTrue())
		})
	})
})
