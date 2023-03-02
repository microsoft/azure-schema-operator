package eventhubs_test

// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
)

var (
	liveTest bool
)

func TestEventhubs(t *testing.T) {
	viper.SetDefault("live_test", false)
	liveTest = viper.GetBool("live_test")
	RegisterFailHandler(Fail)
	RunSpecs(t, "Eventhubs Suite")
}
