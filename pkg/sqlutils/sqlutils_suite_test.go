package sqlutils_test

// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSqlutils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Sqlutils Suite")
}
