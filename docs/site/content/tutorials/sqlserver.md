# SQL Server Tutorial

A common multi-tenancy architecture for sqlserver is using a schema per tenant.
This tutorial will show how to deploy the schema given in a DACPAC format to each schema in the database.

The tutorial uses a [sample dacpac](docs/samples/sqlserver/test.dacpac) with a sales schema.

MSI is used to authenticate in the tutorial as it's simpler.

we need to add the MSI as a user to the DB, e.g.:

```TSQL
CREATE USER [schema-operator-msi] FROM EXTERNAL PROVIDER;
GO
ALTER ROLE db_datareader ADD MEMBER [schema-operator-msi];
ALTER ROLE db_datawriter ADD MEMBER [schema-operator-msi];
ALTER ROLE db_owner ADD MEMBER [schema-operator-msi];
GRANT EXECUTE TO [schema-operator-msi]
GO
```

Creating the ConfigMap:

```bash
kubectl create configmap dacpac-config --from-literal templateName="SalesLT" --from-file=dacpac=docs/samples/sqlserver/test.dacpac
```

next we need to define a `SchemaDeployment` object that will reference the `ConfigMap`.

```yaml
apiVersion: dbschema.microsoft.com/v1alpha1
kind: SchemaDeployment
metadata:
  name: sql-demo-deployment
spec:
  type: sqlServer
  applyTo:
    clusterUris: ['schematest.database.windows.net']
    db: 'db1'
    schema: test
  failIfDataLoss: true
  failurePolicy: abort
  source:
    name: dacpac-config
    namespace: default
```

and apply it via kubectl:

```bash
kubectl apply -f docs/samples/sqlserver/sql-demo-deployment.yaml
```

## External Dacpacs

For cases where the project has external dacpac references we can add them as a reference from the schema ConfigMap like this:

```bash
kubectl create configmap common-config --from-file=dacpac=./DBOCommon.dacpac
kubectl create configmap tenant-config --from-literal templateName="MasterSchema" \
--from-literal externalDacpacs='{ "DBOCommon": {"name": "common-config", "namespace": "default"}}' \
--from-file=dacpac=tenant.dacpac
```

The external Dacpac requires a seperate external `SchemaDeployment` object to deploy it ( to fully capsulate the "externallism" of it)

*Note* as the name of the external DacPac matters we need to pass this name - so it is the "key" for the reference.
