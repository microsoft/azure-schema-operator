# Installation Guide

This document will guide you through the installation and configuration process.
The Schema-Operator needs a client account that can access the Azure Data Explorer clusters you wish to manage
and as this is an Operator we need an AKS cluster.

## Identity

Schema operator uses a managed identity (MSI) to access kusto resources.
Please create a managed identity and assign administrative permissions for the operator to change the databases.  

## Deployment in Dev environment

```bash
export ACR=<your acr>    
export VERSION=0.0.4  
export OPERATOR_IMG="${ACR}.azurecr.io/schema-operator:v${VERSION}"  
az acr login -n ${ACR}
make docker-build-push IMG=$OPERATOR_IMG  
make deploy IMG=$OPERATOR_IMG  
```

## Deployment

Schema-Operator is deployed using a helm chart.
In the provided values we need to pass the MSI name to bind.

```bash
export VERSION=1.0.1  
chart=https://github.com/microsoft/azure-schema-operator/releases/download/v${VERSION}/azure-schema-operator-v${VERSION}.tgz
chart=charts/azure-schema-operator-v${VERSION}.tgz
helm install schema-operator-test $chart --namespace=schema-operator-test --create-namespace
```

more details on chart parameters can be found at the [chart docs](./helm-docs.md)
