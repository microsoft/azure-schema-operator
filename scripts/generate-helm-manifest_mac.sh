#!/bin/bash
# Copyright (c) Microsoft Corporation.
# Licensed under the MIT License.

# This script generates helm manifest and replaces required values in helm chart.

set -x

LOCAL_REGISTRY_CONTROLLER_DOCKER_IMAGE=$1
PUBLIC_REGISTRY=$2
VERSION=$3
DIR=$4

echo "Generating helm chart manifest"
sed -i '' "s@\($PUBLIC_REGISTRY\)\(.*\)@\1azureschemaoperator:$VERSION@g" "$DIR"charts/azure-schema-operator/values.yaml
rm -rf "$DIR"charts/azure-schema-operator/templates/generated
rm -rf "$DIR"charts/azure-schema-operator/crds # remove generated files
mkdir "$DIR"charts/azure-schema-operator/templates/generated
mkdir "$DIR"charts/azure-schema-operator/crds # create dirs for generated files
kustomize build "$DIR"config/default -o "$DIR"charts/azure-schema-operator/templates/generated
find "$DIR"charts/azure-schema-operator/templates/generated/*_customresourcedefinition_* -exec mv '{}' "$DIR"charts/azure-schema-operator/crds \; # move CRD definitions to crd folder
rm "$DIR"charts/azure-schema-operator/templates/generated/*_namespace_* # remove namespace as we will let Helm manage it
sed -i '' "s@$LOCAL_REGISTRY_CONTROLLER_DOCKER_IMAGE@{{.Values.image.repository}}@g" "$DIR"charts/azure-schema-operator/templates/generated/*_deployment_* # Replace hardcoded ASO image
# sed -i '' "s@cert-manager.io/.*@{{.Values.certManagerResourcesAPIVersion}}@g" "$DIR"charts/azure-schema-operator/templates/generated/*cert-manager.io*
find "$DIR"charts/azure-schema-operator/templates/generated/ -type f -exec sed -i '' "s@schema-operator-system@{{.Release.Namespace}}@g" {} \;
sed -i '' "1,/version:.*/s/\(version: \)\(.*\)/\1$VERSION/g" "$DIR"charts/azure-schema-operator/Chart.yaml   # find version key and update the value with the current version
sed -i '' "1s/^/{{- if .Values.ServiceMonitor }} \n/" "$DIR"charts/azure-schema-operator/templates/generated/monitoring.coreos.com_v1_servicemonitor_schema-operator-controller-manager-metrics-monitor.yaml
echo "{{- end }}" >> "$DIR"charts/azure-schema-operator/templates/generated/monitoring.coreos.com_v1_servicemonitor_schema-operator-controller-manager-metrics-monitor.yaml