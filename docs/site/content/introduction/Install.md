---
title: Installation Guide
weight: 1 # want this first
---
# Installation Guide

This document will guide you through the installation and configuration process.
The Schema-Operator needs a client account that can access the Azure Data Explorer clusters you wish to manage
and as this is an Operator we need an AKS cluster.

## Prerequisits

1. An Azure Subscription to provision resources into.
1. An Azure Service Principal for the operator to use. see [Authentication docs](./authentication.md) for further details.
1. A Kubernetes Cluster created and running.

### Identity

Schema operator uses a managed identity (MSI) to access the managed clusters.
Please create a managed identity and assign administrative permissions for the operator to change the databases and schemas.  

## Installation

Schema-Operator is deployed using a helm chart.
In the provided values we need to pass the MSI name to bind.

```bash
export VERSION=0.1.1-alpha 
chart=https://github.com/microsoft/azure-schema-operator/releases/download/v${VERSION}/azure-schema-operator-v${VERSION}.tgz
chart=charts/azure-schema-operator-v${VERSION}.tgz
helm install schema-operator $chart --namespace=schema-operator-namespace --create-namespace
```

more details on chart parameters can be found at the [chart docs](./helm-docs.md)

## Deployment in Dev environment

When developing it's possible to deploy from the repo using `make deploy`  
It will use the aks configured in the local system to deploy the crds and deployment from the `config/` folder.

```bash
export ACR=<your acr>    
export VERSION=0.1.1-alpha 
export OPERATOR_IMG="${ACR}.azurecr.io/schema-operator:v${VERSION}"  
az acr login -n ${ACR}
make docker-build-push IMG=$OPERATOR_IMG  
make deploy IMG=$OPERATOR_IMG  
```

Then install using:

```bash
chart=https://github.com/microsoft/azure-schema-operator/releases/download/v${VERSION}/azure-schema-operator-v${VERSION}.tgz
chart=charts/azure-schema-operator-v${VERSION}.tgz
helm install schema-operator-test $chart --namespace=schema-operator-test --create-namespace --set image.repository=$OPERATOR_IMG
```
