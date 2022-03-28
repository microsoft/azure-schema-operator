# SQL Server Tutorial

A tutorial for a simple scenario of a single db in a server with schema per tenant.

we assume MSI is used to authenticate (because it's simpler :) )

we need to add the MSI as a user to the DB:

```TSQL
CREATE USER [dbset-operator-dataops-msi-erx-qds] FROM EXTERNAL PROVIDER;
GO
ALTER ROLE db_datareader ADD MEMBER [dbset-operator-dataops-msi-erx-qds];
ALTER ROLE db_datawriter ADD MEMBER [dbset-operator-dataops-msi-erx-qds];
ALTER ROLE db_owner ADD MEMBER [dbset-operator-dataops-msi-erx-qds];
GRANT EXECUTE TO [dbset-operator-dataops-msi-erx-qds]
GO
```

Creating the ConfigMap:

```bash
kubectl create configmap dacpac-config --from-literal templateName="SalesLT" --from-file=dacpac=./docs/assets/test.dacpac
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
kubectl apply -f ./docs/assets/sql-demo-deployment.yaml
```

## External Dacpacs

In case our project has external dacpac references we can add them as a reference from the schema ConfigMap:

```bash
kubectl create configmap common-config --from-file=dacpac=./DBOCommon.dacpac
kubectl create configmap tenant-config --from-literal templateName="MasterSchema" \
--from-literal externalDacpacs='{ "DBOCommon": {"name": "common-config", "namespace": "default"}}' \
--from-file=dacpac=./SevilleSqlDbTenant.dacpac
```

The external Dacpac requires a seperate external `SchemaDeployment` object to deploy it ( to fully capsulate the "externallism" of it)

*Note* as the name of the external DacPac matters we need to pass this name - so it is the "key" for the reference.
