package schemaop

// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/cli-runtime/pkg/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"

	schemav1alpha1 "github.com/microsoft/azure-schema-operator/api/v1alpha1"
)

var (
	historyLong = `
		View previous rollout revisions and configurations.`

	historyExample = `
		# View the rollout history of a deployment
		kubectl rollout history deployment/abc
		# View the details of daemonset revision 3
		kubectl rollout history daemonset/abc --revision=3`
)

// SchemaHistoryOptions holds the options for 'schema history' sub command
type SchemaHistoryOptions struct {
	CommonOptions
	configFlags *genericclioptions.ConfigFlags
	PrintFlags  *genericclioptions.PrintFlags
	ToPrinter   func(string) (printers.ResourcePrinter, error)

	Revision int64

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

// NewSchemaHistoryOptions returns an initialized SchemaHistoryOptions instance
func NewSchemaHistoryOptions(streams genericclioptions.IOStreams) *SchemaHistoryOptions {
	o := &SchemaHistoryOptions{
		PrintFlags:  genericclioptions.NewPrintFlags(""),
		configFlags: genericclioptions.NewConfigFlags(true),
		IOStreams:   streams,
	}
	o.SetConfigFlags()
	return o
}

// NewCmdSchemaHistory returns a Command instance for SchemaHistory sub command
func NewCmdSchemaHistory(streams genericclioptions.IOStreams) *cobra.Command {
	o := NewSchemaHistoryOptions(streams)

	cmd := &cobra.Command{
		Use:                   "history (TYPE NAME | TYPE/NAME) [flags]",
		DisableFlagsInUseLine: true,
		Short:                 "View schema history",
		Long:                  historyLong,
		Example:               historyExample,
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

	cmd.Flags().Int64Var(&o.Revision, "revision", o.Revision, "See the details, including podSchemaDeployment of the revision specified")
	cmd.Flags().StringVar(&o.Namespace, "namespace", o.Namespace, "namespace of schema")
	cmd.Flags().StringVar(&o.Name, "name", o.Name, "name of schema template")
	o.PrintFlags.AddFlags(cmd)

	return cmd
}

// Complete completes al the required options
func (o *SchemaHistoryOptions) Complete(cmd *cobra.Command, args []string) error {
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
func (o *SchemaHistoryOptions) Validate() error {
	if o.Revision < 0 {
		return fmt.Errorf("revision must be a positive integer: %v", o.Revision)
	}
	return nil
}

// Run performs the execution of 'rollout history' sub command
func (o *SchemaHistoryOptions) Run() error {

	template := &schemav1alpha1.SchemaDeployment{}
	key := types.NamespacedName{
		Name:      o.Name,
		Namespace: o.Namespace,
	}
	if err := o.Client.Get(context.TODO(), key, template); err != nil {
		return fmt.Errorf("unable to get template: %w", err)
	}
	if o.Revision > 0 {
		revision, _ := o.getRevision(template)
		table := o.newTable([]string{"Namespace", "Name", "Revision", "Executed", "Failed", "Running", "Succeeded"}, o.Out)
		data := []string{revision.Namespace, revision.Name,
			strconv.Itoa(int(revision.Spec.Revision)),
			fmt.Sprintf("%t", revision.Status.Executed),
			strconv.Itoa(int(revision.Status.Failed)),
			strconv.Itoa(int(revision.Status.Running)),
			strconv.Itoa(int(revision.Status.Succeeded))}
		table.Append(data)
		table.Render()
	} else {
		table := o.newTable([]string{"Namespace", "Name", "Revision"}, o.Out)
		vdList, _ := o.getVersionedDeployments(template)
		for _, item := range vdList {
			data := []string{item.Namespace, item.Name}
			data = append(data, strconv.Itoa(int(item.Spec.Revision)))
			// fmt.Printf("%d) got resource: %s, namespace: %s, revision: %d \n", i, item.Name, item.Namespace, item.Spec.Revision)
			table.Append(data)
		}
		// Send output.
		table.Render()

	}
	return nil
}

func (o *SchemaHistoryOptions) getVersionedDeployments(template *schemav1alpha1.SchemaDeployment) ([]schemav1alpha1.VersionedDeplyment, error) {
	vdList := &schemav1alpha1.VersionedDeplymentList{}
	if err := o.Client.List(context.TODO(), vdList, &client.ListOptions{Namespace: o.Namespace}); err != nil {
		return nil, fmt.Errorf("unable to list versioned deployments: %w", err)
	}
	// Only include those whose ControllerRef matches the Deployment.
	owned := make([]schemav1alpha1.VersionedDeplyment, 0, len(vdList.Items))
	for _, vd := range vdList.Items {
		if metav1.IsControlledBy(&vd, template) {
			// fmt.Printf("owned resource: %s, namespace: %s, revision: %d \n", vd.Name, vd.Namespace, vd.Spec.Revision)
			owned = append(owned, vd)
		}
	}
	return owned, nil
}

func (o *SchemaHistoryOptions) getRevision(template *schemav1alpha1.SchemaDeployment) (*schemav1alpha1.VersionedDeplyment, error) {
	revisionDeployment := &schemav1alpha1.VersionedDeplyment{}
	key := types.NamespacedName{
		Name:      template.Name + "-" + strconv.Itoa(int(o.Revision)),
		Namespace: o.Namespace,
	}
	if err := o.Client.Get(context.TODO(), key, revisionDeployment); err != nil {
		return nil, fmt.Errorf("unable to list versioned deployments: %w", err)
	}

	return revisionDeployment, nil
}
