# Azure Data Explorer (ADX, AKA Kusto) Tutorial

In this short tutorial we will review the process of managing a schema, performing a change and rollback in the case of an error.
We will deploy 3 revisions of our schema, with an error on the third schema triggering a rollback.

The schema is represented in `kql` field in a standard `ConfigMap` which contains ADX schemas described as KQLs..
More details on KQL files can be found in the [delta-kusto instructions](https://github.com/microsoft/delta-kusto/blob/main/documentation/tutorials/overview-tutorial/README.md#download-dev)  
Or simply download and review a [sample kql file](https://github.com/microsoft/delta-kusto/blob/main/documentation/tutorials/overview-tutorial/dev-start-samples.kql)

Once we have a kql describe our schema we can generate a `ConfigMap` using:

```sh
kubectl create configmap test-sample-kql --from-file=kql=sample.kql --dry-run=client -o yaml
```

We reference the ConfigMap object from the `SchemaDeployment` object to apply onto the clusters.

All The objects used throughout the tutorial can be found in [samples folder](docs/samples/kusto)

## Pre-requisits

the tutorial assumes that the Schema operator is already installed with the appropriate permissions - if not, please see [installation](content/introduction/Install.md)
While not mandetory, a kubectl plugin exists that provides simpler access to the schema revision history - see [plugin installation](content/introduction/plugin_installation.md) for details

## Tutorial steps

The first step is to create our first schema `ConfigMap`, later to be deployed to our cluster:

```sh
kubectl apply -f docs/samples/kusto/sample-cm.yml
```

With the schema's `ConfigMap` in place we are ready to deploy to our test cluster:

```sh
kubectl apply -f docs/samples/kusto/sample-sd.yml
```

Once we've created the necessary k8s objects, we should check the deployment status by getting the `SchemaDeployment` object:

```bash
➜ kubectl get schemadeployments sample-adx
NAME         TYPE    EXECUTED
sample-adx   kusto   True
```

To update the schema, we should simply patch the ConfigMap, which will trigger the schema operator to validate the schema, and apply updates if needed.

```sh
kubectl apply -f docs/samples/kusto/sample-cm-v2.yml  
```

To view current status and history we can use the provided plugin:

```sh
kubectl schemaop status --namespace default --name sample-adx     
  NAMESPACE  NAME          REVISION  EXECUTED  FAILED  RUNNING  SUCCEEDED  
  default    sample-adx-1  1         true      0       0        1          

kubectl schemaop history --namespace default --name sample-adx    
  NAMESPACE  NAME          REVISION  
  default    sample-adx-0  0         
  default    sample-adx-1  1      
```

Now, lets make things a bit more interesting, and apply a schema that contains an error:

```sh
kubectl apply -f docs/samples/kusto/sample-cm-err.yml
```

If we check the status we will see we are now on our 4th revision!

```sh
kubectl schemaop history --namespace default --name sample-adx              

  NAMESPACE  NAME          REVISION  
  default    sample-adx-0  0         
  default    sample-adx-1  1         
  default    sample-adx-2  2         
  default    sample-adx-3  3         
```

Further checking the history we can see that revision 2 (the 3rd revision) failed:

```sh
kubectl schemaop history --namespace default --name sample-adx  --revision 2
  NAMESPACE  NAME          REVISION  EXECUTED  FAILED  RUNNING  SUCCEEDED  
  default    sample-adx-2  2         false     1       1        0     
```

We can also see, that the 4th and 2nd revisions are the same.
That's because we rolled back the faulty schema revision as requested by the template object:

```yaml
spec:
  failurePolicy: rollback
```

## Summary

We've looked at a common flow and how Schema-Operator can ease managment of large schema deployments.
We suggest trying some of these scenarios in a dev environment to get familiar with the different options.
