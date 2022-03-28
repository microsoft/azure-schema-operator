package schemaop

// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/cli-runtime/pkg/resource"

	schemav1alpha1 "github.com/microsoft/azure-schema-operator/api/v1alpha1"
)

var (
	statusLong = `
		View schema rollout status.`

	statusExample = `
		# View the schema rollout status
		kubectl schemaop status --name master-test-template`
)

// SchemaStatusOptions holds the options for 'schema history' sub command
type SchemaStatusOptions struct {
	CommonOptions
	configFlags *genericclioptions.ConfigFlags
	PrintFlags  *genericclioptions.PrintFlags
	ToPrinter   func(string) (printers.ResourcePrinter, error)

	Builder          *resource.Builder
	Resources        []string
	Namespace        string
	Name             string
	EnforceNamespace bool
	RESTClientGetter genericclioptions.RESTClientGetter
	// userSpecifiedNamespace string

	resource.FilenameOptions
	genericclioptions.IOStreams
}

// NewSchemaStatusOptions returns an initialized SchemaStatusOptions instance
func NewSchemaStatusOptions(streams genericclioptions.IOStreams) *SchemaStatusOptions {
	o := &SchemaStatusOptions{
		PrintFlags:  genericclioptions.NewPrintFlags(""),
		configFlags: genericclioptions.NewConfigFlags(true),
		IOStreams:   streams,
	}
	o.SetConfigFlags()
	return o
}

// NewCmdSchemaStatus returns a Command instance for status sub command
func NewCmdSchemaStatus(streams genericclioptions.IOStreams) *cobra.Command {
	o := NewSchemaStatusOptions(streams)

	cmd := &cobra.Command{
		Use:                   "status (TYPE NAME | TYPE/NAME) [flags]",
		DisableFlagsInUseLine: true,
		Short:                 "View schema rollout status",
		Long:                  statusLong,
		Example:               statusExample,
		RunE: func(c *cobra.Command, args []string) error {
			if err := o.Complete(c, args); err != nil {
				return err
			}
			if err := o.Validate(); err != nil {
				return err
			}
			if err := o.Run(); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&o.Namespace, "namespace", o.Namespace, "namespace of schema")
	cmd.Flags().StringVar(&o.Name, "name", o.Name, "name of schema template")
	o.PrintFlags.AddFlags(cmd)

	return cmd
}

// Complete completes al the required options
func (o *SchemaStatusOptions) Complete(cmd *cobra.Command, args []string) error {
	o.Resources = args

	var err error
	o.Namespace, err = cmd.Flags().GetString("namespace")
	if err != nil {
		return err
	}

	o.Name, err = cmd.Flags().GetString("name")
	if err != nil {
		return err
	}

	o.ToPrinter = func(operation string) (printers.ResourcePrinter, error) {
		o.PrintFlags.NamePrintFlags.Operation = operation
		return o.PrintFlags.ToPrinter()
	}

	return o.Init(cmd)

}

// Validate makes sure all the provided values for command-line options are valid
func (o *SchemaStatusOptions) Validate() error {

	return nil
}

// Run performs the execution of 'rollout history' sub command
func (o *SchemaStatusOptions) Run() error {

	template := &schemav1alpha1.SchemaDeployment{}
	key := types.NamespacedName{
		Name:      o.Name,
		Namespace: o.Namespace,
	}
	if err := o.Client.Get(context.TODO(), key, template); err != nil {
		return fmt.Errorf("unable to get template: %w", err)
	}
	revision, _ := o.getCurrentRevision(template)
	table := o.newTable([]string{"Namespace", "Name", "Revision", "Executed", "Failed", "Running", "Succeeded"}, o.Out)
	data := []string{revision.Namespace, revision.Name,
		strconv.Itoa(int(revision.Spec.Revision)),
		fmt.Sprintf("%t", revision.Status.Executed),
		strconv.Itoa(int(revision.Status.Failed)),
		strconv.Itoa(int(revision.Status.Running)),
		strconv.Itoa(int(revision.Status.Succeeded))}
	table.Append(data)
	table.Render()

	return nil
}

func (o *SchemaStatusOptions) getCurrentRevision(template *schemav1alpha1.SchemaDeployment) (*schemav1alpha1.VersionedDeplyment, error) {
	revisionDeployment := &schemav1alpha1.VersionedDeplyment{}
	key := types.NamespacedName{
		Name:      template.Status.CurrentVerDeployment.Name,
		Namespace: template.Status.CurrentVerDeployment.Namespace,
	}
	if err := o.Client.Get(context.TODO(), key, revisionDeployment); err != nil {
		return nil, fmt.Errorf("unable to list versioned deployments: %w", err)
	}

	return revisionDeployment, nil
}
