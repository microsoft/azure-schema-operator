package kustoutils_test

// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/microsoft/azure-schema-operator/pkg/kustoutils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type Filtered struct {
	DBS []string `json:"dbs"`
}

var _ = Describe("Server-side Request handling", func() {
	var (
		c       *kustoutils.WebHookClient
		handler http.Handler
		srv     *httptest.Server
	)

	BeforeEach(func() {
		handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Expect(r.URL.Path).To(Equal("/resource"))
			query := r.URL.Query()
			cluster, ok := query["cluster"]
			if !ok {
				fmt.Println("cluster not provided")
				w.WriteHeader(500)
				return
			}

			label, present := query["label"] //filters=["color", "price", "brand"]
			if !present {
				fmt.Println("filters not present")
				return
			}
			fmt.Printf("getting dbs for cluster %s, label: %s\n", cluster, label)
			ret := Filtered{}
			if label[0] == "delux" {
				ret.DBS = []string{"db1937", "db2022"}
			} else {
				ret.DBS = []string{"db1938", "db2020"}
			}

			w.WriteHeader(200)
			w.Header().Set("Content-Type", "application/json")
			b, _ := json.Marshal(ret)
			_, _ = w.Write(b)
		})
	})

	JustBeforeEach(func() {
		srv = httptest.NewServer(handler)
		c = kustoutils.NewWebHookClient(srv.Client())
	})

	AfterEach(func() {
		srv.CloseClientConnections()
		srv.Close()
	})

	It("Checks the resource path", func() {
		url := srv.URL + "/dbs?cluster={{.Cluster}}&label={{.Label}}"
		dbs, err := c.PerformQuery(url, "test-cluster", "delux")
		Expect(err).ToNot(HaveOccurred())
		Expect(len(dbs)).To(Equal(2))
	})

	// Context("Use a different handler", func() {
	// 	BeforeEach(func() {
	// 		handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 			Expect(r.Method).To(Equal(http.MethodPost))
	// 			w.WriteHeader(http.StatusCreated)
	// 		})
	// 	})

	// 	It("Checks the request method", func() {
	// 		c.PerformQuery()
	// 	})
	// })
})
