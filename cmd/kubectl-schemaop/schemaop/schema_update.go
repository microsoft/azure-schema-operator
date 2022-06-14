package schemaop

// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/cli-runtime/pkg/resource"
)

var (
	updateLong = `
		update the schema configMap.`

	updateExample = `
		# View the schema rollout status
		kubectl schemaop update --name master-test-template --schema-file /path/to/schema/file`
)

// SchemaUpdateOptions holds the options for 'schema history' sub command
type SchemaUpdateOptions struct {
	CommonOptions
	configFlags *genericclioptions.ConfigFlags
	PrintFlags  *genericclioptions.PrintFlags
	ToPrinter   func(string) (printers.ResourcePrinter, error)

	Builder          *resource.Builder
	Resources        []string
	Namespace        string
	Name             string
	SchemaFile       string
	EnforceNamespace bool
	DryRun           bool
	RESTClientGetter genericclioptions.RESTClientGetter
	// userSpecifiedNamespace string

	resource.FilenameOptions
	genericclioptions.IOStreams
}

// NewSchemaUpdateOptions returns an initialized SchemaUpdateOptions instance
func NewSchemaUpdateOptions(streams genericclioptions.IOStreams) *SchemaUpdateOptions {
	o := &SchemaUpdateOptions{
		PrintFlags:  genericclioptions.NewPrintFlags(""),
		configFlags: genericclioptions.NewConfigFlags(true),
		IOStreams:   streams,
	}
	o.SetConfigFlags()
	return o
}

// NewCmdSchemaUpdate returns a Command instance for update sub command
func NewCmdSchemaUpdate(streams genericclioptions.IOStreams) *cobra.Command {
	o := NewSchemaUpdateOptions(streams)

	cmd := &cobra.Command{
		Use:                   "update (TYPE NAME | TYPE/NAME) [flags]",
		DisableFlagsInUseLine: true,
		Short:                 "update schema configMap",
		Long:                  updateLong,
		Example:               updateExample,
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
	cmd.Flags().StringVar(&o.Name, "name", o.Name, "name of schema configMap")
	cmd.Flags().StringVar(&o.SchemaFile, "schema-file", o.SchemaFile, "path to the schema file")
	cmd.Flags().BoolVar(&o.DryRun, "dry-run", o.DryRun, "dry-run - only print")
	o.PrintFlags.AddFlags(cmd)

	return cmd
}

// Complete completes al the required options
func (o *SchemaUpdateOptions) Complete(cmd *cobra.Command, args []string) error {
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

	o.SchemaFile, err = cmd.Flags().GetString("schema-file")
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
func (o *SchemaUpdateOptions) Validate() error {
	// check that the given schema file path exists
	_, err := os.Stat(o.SchemaFile)
	return err
}

// Run performs the execution of 'rollout history' sub command
func (o *SchemaUpdateOptions) Run() error {

	found := false
	sourceCfgMap := &v1.ConfigMap{}
	key := types.NamespacedName{
		Name:      o.Name,
		Namespace: o.Namespace,
	}
	if err := o.Client.Get(context.TODO(), key, sourceCfgMap); err != nil {
		fmt.Println("unable to find schema config map - will create")

		sourceCfgMap.ObjectMeta.Name = o.Name
		sourceCfgMap.ObjectMeta.Namespace = o.Namespace
		sourceCfgMap.Data = make(map[string]string)
	} else {
		found = true
	}
	b, err := ioutil.ReadFile(o.SchemaFile) // just pass the file name
	if err != nil {
		return fmt.Errorf("unable to read schema file: %w", err)
	}
	sourceCfgMap.Data["kql"] = string(b)

	if o.DryRun {
		out, err := yaml.Marshal(sourceCfgMap)
		if err != nil {
			return fmt.Errorf("failed to unmarshel configMap: %w", err)
		}
		fmt.Print(string(out))
	} else {
		if found {
			err = o.Client.Update(context.Background(), sourceCfgMap)
		} else {
			err = o.Client.Create(context.Background(), sourceCfgMap)
		}

		if err != nil {
			return fmt.Errorf("unable to update source configMap: %w", err)
		}

	}
	fmt.Println("schema config map updated with file content")
	return nil
}
