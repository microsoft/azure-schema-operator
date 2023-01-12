package schemaop

// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/cli-runtime/pkg/resource"

	schemav1alpha1 "github.com/microsoft/azure-schema-operator/apis/dbschema/v1alpha1"
	"github.com/microsoft/azure-schema-operator/pkg/utils/schemaversions"
)

var (
	undoLong = `
		undo previous rollout revisions.`

	undoExample = `
		# revert back to revision 3
		kubectl schemaop undo daemonset/abc --to-revision=3`
)

// SchemaUndoOptions holds the options for 'schema history' sub command
type SchemaUndoOptions struct {
	CommonOptions
	configFlags *genericclioptions.ConfigFlags
	PrintFlags  *genericclioptions.PrintFlags
	ToPrinter   func(string) (printers.ResourcePrinter, error)

	ToRevision int32

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

// NewSchemaUndoOptions returns an initialized SchemaUndoOptions instance
func NewSchemaUndoOptions(streams genericclioptions.IOStreams) *SchemaUndoOptions {
	o := &SchemaUndoOptions{
		PrintFlags:  genericclioptions.NewPrintFlags(""),
		configFlags: genericclioptions.NewConfigFlags(true),
		IOStreams:   streams,
		ToRevision:  int32(0),
	}
	o.SetConfigFlags()
	return o
}

// NewCmdSchemaUndo returns a Command instance for SchemaUndo sub command
func NewCmdSchemaUndo(streams genericclioptions.IOStreams) *cobra.Command {
	o := NewSchemaUndoOptions(streams)

	cmd := &cobra.Command{
		Use:                   "undo (TYPE NAME | TYPE/NAME) [flags]",
		DisableFlagsInUseLine: true,
		Short:                 "undo schema rollout",
		Long:                  undoLong,
		Example:               undoExample,
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

	cmd.Flags().Int32Var(&o.ToRevision, "to-revision", o.ToRevision, "The revision to rollback to. Default to 0 (last revision).")
	cmd.Flags().StringVar(&o.Namespace, "namespace", o.Namespace, "namespace of schema")
	cmd.Flags().StringVar(&o.Name, "name", o.Name, "name of schema template")
	o.PrintFlags.AddFlags(cmd)

	return cmd
}

// Complete completes al the required options
func (o *SchemaUndoOptions) Complete(cmd *cobra.Command, args []string) error {
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
func (o *SchemaUndoOptions) Validate() error {
	if o.ToRevision < 0 {
		return fmt.Errorf("revision must be a positive integer: %v", o.ToRevision)
	}
	return nil
}

// Run performs the execution of 'rollout history' sub command
func (o *SchemaUndoOptions) Run() error {

	template := &schemav1alpha1.SchemaDeployment{}
	key := types.NamespacedName{
		Name:      o.Name,
		Namespace: o.Namespace,
	}
	if err := o.Client.Get(context.TODO(), key, template); err != nil {
		return fmt.Errorf("unable to get template: %w", err)
	}
	if o.ToRevision > 0 {
		err := schemaversions.RollbackToVersion(o.Client, template, o.ToRevision)
		return err
	}
	return nil
}

// func (o *SchemaUndoOptions) getVersionedDeployments(template *schemav1alpha1.SchemaDeployment) ([]schemav1alpha1.VersionedDeplyment, error) {
// 	vdList := &schemav1alpha1.VersionedDeplymentList{}
// 	if err := o.Client.List(context.TODO(), vdList, &client.ListOptions{Namespace: o.Namespace}); err != nil {
// 		return nil, fmt.Errorf("unable to list versioned deployments: %w", err)
// 	}
// 	// Only include those whose ControllerRef matches the Deployment.
// 	owned := make([]schemav1alpha1.VersionedDeplyment, 0, len(vdList.Items))
// 	for _, vd := range vdList.Items {
// 		if metav1.IsControlledBy(&vd, template) {
// 			// fmt.Printf("owned resource: %s, namespace: %s, revision: %d \n", vd.Name, vd.Namespace, vd.Spec.Revision)
// 			owned = append(owned, vd)
// 		}
// 	}
// 	return owned, nil
// }

// func (o *SchemaUndoOptions) getRevision(template *schemav1alpha1.SchemaDeployment) (*schemav1alpha1.VersionedDeplyment, error) {
// 	revisionDeployment := &schemav1alpha1.VersionedDeplyment{}
// 	key := types.NamespacedName{
// 		Name:      template.Name + "-" + strconv.Itoa(int(o.ToRevision)),
// 		Namespace: o.Namespace,
// 	}
// 	if err := o.Client.Get(context.TODO(), key, revisionDeployment); err != nil {
// 		return nil, fmt.Errorf("unable to list versioned deployments: %w", err)
// 	}

// 	return revisionDeployment, nil
// }
