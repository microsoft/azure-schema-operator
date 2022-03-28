package schemaop

// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
import (
	"github.com/spf13/cobra"

	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var (
	rolloutLong = `	Manage the rollout of a resource.`

	rolloutExample = `
		# Rollback to the previous deployment
		kubectl rollout undo deployment/abc
		# Check the rollout status of a daemonset
		kubectl rollout status daemonset/foo`

	// rolloutValidResources = dedent.Dedent(`
	// 	Valid resource types include:
	// 	   * deployments
	// 	   * daemonsets
	// 	   * statefulsets
	// 	`)
)

// NewCmd returns a Command instance for 'schemaop' sub command
func NewCmd(streams genericclioptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "schemaop SUBCOMMAND",
		DisableFlagsInUseLine: true,
		Short:                 "Manage the rollout of a schema",
		Long:                  rolloutLong,
		Example:               rolloutExample,
	}
	// subcommands
	cmd.AddCommand(NewCmdSchemaHistory(streams))
	cmd.AddCommand(NewCmdSchemaStatus(streams))
	// cmd.AddCommand(NewCmdRolloutPause(f, streams))
	// cmd.AddCommand(NewCmdRolloutResume(f, streams))
	cmd.AddCommand(NewCmdSchemaUndo(streams))
	cmd.AddCommand(NewCmdSchemaUpdate(streams))
	// cmd.AddCommand(NewCmdRolloutStatus(f, streams))
	// cmd.AddCommand(NewCmdRolloutRestart(f, streams))

	return cmd
}
