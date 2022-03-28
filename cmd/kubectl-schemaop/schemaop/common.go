package schemaop

// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
import (
	"fmt"
	"io"

	schemav1alpha1 "github.com/microsoft/azure-schema-operator/api/v1alpha1"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

// CommonOptions Options encapsulates the common fields of command options
type CommonOptions struct {
	ConfigFlags   *genericclioptions.ConfigFlags
	Client        client.Client
	Clientset     *kubernetes.Clientset
	UserNamespace string
}

// Init initialize the common config of command options
func (o *CommonOptions) Init(cmd *cobra.Command) error {
	clientConfig := o.GetClientConfig()

	client, err := NewClient(clientConfig)
	if err != nil {
		return fmt.Errorf("unable to instantiate client: %w", err)
	}
	o.SetClient(client)

	clientset, err := NewClientset(clientConfig)
	if err != nil {
		return fmt.Errorf("unable to instantiate clientset: %w", err)
	}
	o.SetClientset(clientset)

	nsConfig, _, err := clientConfig.Namespace()
	if err != nil {
		return err
	}

	nsFlag, err := cmd.Flags().GetString("namespace")
	if err != nil {
		return err
	}

	if nsFlag != "" {
		o.SetNamespace(nsFlag)
	} else {
		o.SetNamespace(nsConfig)
	}

	return nil
}

// SetNamespace configures the namespace
func (o *CommonOptions) SetNamespace(ns string) {
	o.UserNamespace = ns
}

// SetClient configures the client
func (o *CommonOptions) SetClient(client client.Client) {
	o.Client = client
}

// SetClientset configures the clientset
func (o *CommonOptions) SetClientset(clientset *kubernetes.Clientset) {
	o.Clientset = clientset
}

// GetClientConfig returns the client config
func (o *CommonOptions) GetClientConfig() clientcmd.ClientConfig {
	return o.ConfigFlags.ToRawKubeConfigLoader()
}

// SetConfigFlags configures the config flags
func (o *CommonOptions) SetConfigFlags() {
	o.ConfigFlags = genericclioptions.NewConfigFlags(false)
}

// sample []string{"Namespace", "Name", "Agent", "Cluster-Agent", "Cluster-Checks-Runner", "Age"}
func (o *CommonOptions) newTable(headers []string, out io.Writer) *tablewriter.Table {
	table := tablewriter.NewWriter(out)
	table.SetHeader(headers)
	table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetRowLine(false)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeaderLine(false)
	return table
}

// NewClient returns a new controller-runtime client instance
func NewClient(clientConfig clientcmd.ClientConfig) (client.Client, error) {
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("unable to get rest client config: %w", err)
	}

	// Create the mapper provider
	mapper, err := apiutil.NewDiscoveryRESTMapper(restConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to to instantiate mapper: %w", err)
	}

	// Register the scheme
	if err = schemav1alpha1.AddToScheme(scheme.Scheme); err != nil {
		return nil, fmt.Errorf("unable register DatadogAgent apis: %w", err)
	}

	// Create the Client for Read/Write operations.
	var newClient client.Client
	newClient, err = client.New(restConfig, client.Options{Scheme: scheme.Scheme, Mapper: mapper})
	if err != nil {
		return nil, fmt.Errorf("unable to instantiate client: %w", err)
	}

	return newClient, nil
}

// NewClientset returns a new client-go instance
func NewClientset(clientConfig clientcmd.ClientConfig) (*kubernetes.Clientset, error) {
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("unable to get rest client config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to instantiate client: %w", err)
	}

	return clientset, nil
}
