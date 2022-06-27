package kustoutils_test

// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
import (
	"strings"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
)

var (
	liveTest    bool
	testCluster = strings.TrimSpace(viper.GetString("schemaop_test_kusto_cluster_name"))
)

func TestKustoutils(t *testing.T) {
	viper.SetDefault("live_test", false)
	liveTest = viper.GetBool("live_test")
	RegisterFailHandler(Fail)
	RunSpecs(t, "Kustoutils Suite")
}
