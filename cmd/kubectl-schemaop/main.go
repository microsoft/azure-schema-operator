package main

// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
import (
	"os"

	"github.com/microsoft/azure-schema-operator/cmd/kubectl-schemaop/schemaop"

	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func main() {
	flags := pflag.NewFlagSet("kubectl-schemaop", pflag.ExitOnError)
	pflag.CommandLine = flags

	root := schemaop.NewCmd(genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr})
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
