module github.com/microsoft/azure-schema-operator

go 1.16

require (
	github.com/Azure/azure-kusto-go v0.4.1
	github.com/Azure/azure-sdk-for-go/sdk/azcore v0.19.0
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v0.11.0
	github.com/Azure/go-autorest/autorest v0.11.24
	github.com/Azure/go-autorest/autorest/adal v0.9.18
	github.com/Azure/go-autorest/autorest/azure/auth v0.4.2
	github.com/Azure/go-autorest/autorest/validation v0.3.1
	github.com/Azure/go-autorest/tracing v0.6.0
	github.com/denisenkom/go-mssqldb v0.12.0
	github.com/go-logr/logr v0.4.0
	github.com/lithammer/dedent v1.1.0
	github.com/olekukonko/tablewriter v0.0.5
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.16.0
	github.com/prometheus/client_golang v1.11.0
	github.com/rs/zerolog v1.25.0
	github.com/spf13/cobra v1.2.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.9.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	k8s.io/api v0.22.3
	k8s.io/apimachinery v0.22.3
	k8s.io/cli-runtime v0.22.3
	k8s.io/client-go v0.22.3
	sigs.k8s.io/controller-runtime v0.9.2
)
