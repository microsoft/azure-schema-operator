package eventhubs_test

// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
import (
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"

	schemav1alpha1 "github.com/microsoft/azure-schema-operator/apis/dbschema/v1alpha1"
	"github.com/microsoft/azure-schema-operator/pkg/eventhubs"
	"github.com/microsoft/azure-schema-operator/pkg/eventhubs/azure/schemaregistry"
)

var _ = Describe("Schemaregistery", func() {
	Context("when creating configMap", func() {
		targets := schemav1alpha1.ClusterTargets{}
		cfgMap := &v1.ConfigMap{}
		cfgMap.Data = make(map[string]string)
		cfgMap.Data["templateName"] = "schemaop"
		cfgMap.Data["group"] = "testsgr"
		cfgMap.Data["schema"] = `{"name":"schemaop","namespace":"com.azure.schemaregistry.samples","type":"record","fields":[{"name":"id","type":"string"},{"name":"amount","type":"double"}]}`
		schema := schemaregistry.Schema{
			Name:      "schemaop",
			Namespace: "com.azure.schemaregistry.samples",
			Type:      "record",
			Fields: []interface{}{
				map[string]interface{}{
					"name": "id",
					"type": "string",
				},
				map[string]interface{}{
					"name": "amount",
					"type": "double",
				},
			},
		}
		schemaContent, err := json.Marshal(schema)
		Expect(err).NotTo(HaveOccurred())

		config := schemav1alpha1.ExecutionConfiguration{
			Group:        "testsgr",
			TemplateName: "schemaop",
			Schema:       string(schemaContent),
		}
		if liveTest {
			It("Should parse and extract configuration from configMap", func() {
				registry := eventhubs.NewRegistry("jonytest.servicebus.windows.net")
				ec, err := registry.CreateExecConfiguration(targets, cfgMap, false)
				Expect(err).NotTo(HaveOccurred())
				Expect(ec).To(Equal(config))
			})
			It("It Should register the schema", func() {
				registry := eventhubs.NewRegistry("jonytest.servicebus.windows.net")
				_, err = registry.Execute(targets, config)
				Expect(err).NotTo(HaveOccurred())
			})
		}

	})

})
