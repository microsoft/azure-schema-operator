package kustoutils_test

// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
import (
	"fmt"

	"github.com/microsoft/azure-schema-operator/pkg/kustoutils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DeltaWrapper", func() {
	Context("when creating configuration", func() {
		It("Should generate configuration", func() {
			w := kustoutils.NewDeltaWrapper()
			uri := "https://sampleadx.westeurope.kusto.windows.net"
			dbs := []string{"db1", "db2", "db3"}
			kqlFile := "/Users/jocohe/Documents/delta-kusto/dev-state.kql"
			fileName, err := w.CreateExecConfiguration(uri, dbs, kqlFile, true)
			Expect(err).NotTo(HaveOccurred())
			fmt.Fprintf(GinkgoWriter, "generated config file: %s\n", fileName)
		})

	})
})
