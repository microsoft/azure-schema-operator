package schemaop

// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version returns the operator CLI version
// TODO: get the build version
func Version() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "version",
		Short:         "schema-operator version information",
		Long:          `Prints the current version of the schema-operator CLI. This may or may not match the version in the cluster.`,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("schema-operator %s\n", "test-0.0.1")
			return nil
		},
	}

	return cmd
}
