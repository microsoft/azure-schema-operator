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

Schema-Operator is deployed using a helm chart. In the provided values we need to pass the MSI name to bind.

Helm (helm3 upgrade schema-operator-test azure-schema-operator --install --set namespace=wdatp-infra-system --dry-run --timeout 60s --debug)
