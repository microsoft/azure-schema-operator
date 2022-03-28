package kustoutils_test

// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestKustoutils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Kustoutils Suite")
}
