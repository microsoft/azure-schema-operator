# Azure Data Explorer (ADX, AKA Kusto) Tutorial

In this short tutorial we will review the process of managing a schema, performing a change and rollback in the case of an error.

## Pre-requisits

the tutorial assumes that the Schema operator is already installed with the appropriate permissions - if not, please see [installation](Install.md)

## Tutorial steps

We start by creating a KQL file following the [delta-kusto instructions](https://github.com/microsoft/delta-kusto/blob/main/documentation/tutorials/overview-tutorial/README.md#download-dev)  
Or simply download a [sample kql file](https://github.com/microsoft/delta-kusto/blob/main/documentation/tutorials/overview-tutorial/dev-start-samples.kql)

Once we have a kql file we need to generate a `ConfigMap` definition file:

```sh
kubectl create configmap test-sample-kql --from-file=kql=/Users/jocohe/Documents/delta-kusto/sample.kql --dry-run=client -o json | jq .
```

We reference the ConfigMap from our `SchemaDeployment` object (name to be changed...) to apply to our clusters

```sh
kubectl apply -f /Users/jocohe/Documents/delta-kusto/template-demo.yml
```

To update the schema to a new version we can simply apply a new ConfigMap:

```sh
kubectl apply -f /Users/jocohe/Documents/delta-kusto/cm-dev2.yml  
```

or use the kuebctl `schemaop` plugin:

```sh
kubectl schemaop update --namespace default --name dev-test-kql --schema-file /Users/jocohe/Documents/delta-kusto/dev-state.kql
```

To view current status and history we can use the provided plugin:

```sh
kubectl schemaop status --namespace default --name master-test-template     
  NAMESPACE  NAME                    REVISION  EXECUTED  FAILED  RUNNING  SUCCEEDED  
  default    master-test-template-1  1         true      0       0        1          

kubectl schemaop history --namespace default --name master-test-template    
  NAMESPACE  NAME                    REVISION  
  default    master-test-template-0  0         
  default    master-test-template-1  1      
```

Now, lets make things a bit more interesting, and apply a bad schema:

```sh
kubectl apply -f /Users/jocohe/Documents/delta-kusto/cm-dev-err.yml
```

If we check the status we will see we are now on our 4th revision!

```sh
kubectl schemaop history --namespace default --name master-test-template              

  NAMESPACE  NAME                    REVISION  
  default    master-test-template-0  0         
  default    master-test-template-1  1         
  default    master-test-template-2  2         
  default    master-test-template-3  3         
```

further checking the history we can see that revisio 2 (the 3rd revision) failed:

```sh
kubectl schemaop history --namespace default --name master-test-template  --revision 2
  NAMESPACE  NAME                    REVISION  EXECUTED  FAILED  RUNNING  SUCCEEDED  
  default    master-test-template-2  2         false     1       1        0     
```

and we can see that the 4th and 2nd revisions are the same - we rolled back the bad revision as requested by the template object:

```yaml
spec:
  failurePolicy: rollback
```

## Summary

We've looked at a common flow and how Schema-Operator can ease managment of large schema deployments.
We suggest trying some of these scenarios in a dev environment to get familiar with the tool.
