package sqlutils_test

// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
import (
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
)

var (
	liveTest    bool
	testCluster = strings.TrimSpace(viper.GetString("schemaop_test_sqlserver_cluster_name"))
)

func TestSqlutils(t *testing.T) {
	viper.SetDefault("live_test", false)
	liveTest = viper.GetBool("live_test")
	RegisterFailHandler(Fail)
	RunSpecs(t, "Sqlutils Suite")
}
