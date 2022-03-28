# Azure Schema Registry for Go

> see <https://aka.ms/autorest>

### Generation

```ps
autorest --reset
autorest --low-level-client --modelerfour.lenient-model-deduplication README.md
```

### Settings

``` yaml
input-file: https://raw.githubusercontent.com/Azure/azure-rest-api-specs/main/specification/schemaregistry/data-plane/Microsoft.EventHub/stable/2021-10/schemaregistry.json
output-folder: ./schemaregistry
namespace: schemaregistry
no-namespace-folders: true
license-header: MICROSOFT_MIT_NO_VERSION
clear-output-folder: true
go: true
add-credential: true
credential-scopes: "https://eventhubs.azure.net/.default"
package-version: "1.0.0"
```

**Note:** This is the base client code generation - it was changed.
this is left for historical reasons and documentation
