#!/usr/bin/env bash

# Copyright (c) Microsoft Corporation.
# Licensed under the MIT license.

set -o errexit
set -o nounset
set -o pipefail

# Enable tracing in this script off by setting the TRACE variable in your
# environment to any value:
#
# $ TRACE=1 test.sh
TRACE=${TRACE:-""}
if [[ -n "${TRACE}" ]]; then
  set -x
fi

REPO_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
subscription_id=$(az account show -o json | jq -r ".id")
json=$(az ad sp create-for-rbac -n "schema-operator-tester" -o json)
client_id=$(echo "$json" | jq -r '.appId')
client_secret=$(echo "$json" | jq -r '.password')
tenant_id=$(echo "$json" | jq -r '.tenant')


echo -e "AZURE_TENANT_ID=$tenant_id\nAZURE_CLIENT_ID=$client_id\nAZURE_CLIENT_SECRET=$client_secret\nAZURE_SUBSCRIPTION_ID=$subscription_id" > "$REPO_ROOT"/.env

az kusto cluster-principal-assignment create \
--cluster-name $SCHEMAOP_TEST_KUSTO_CLUSTER_NAME \
--principal-id $client_id \
--principal-type "App" \
--role "AllDatabasesAdmin" \
--tenant-id $tenant_id \
--principal-assignment-name "schemaopprincipal" \
--resource-group $SCHEMAOP_TEST_RG_NAME

echo """
Run the following on the SQL server to enable access for the test cluster:
CREATE USER [schema-operator-tester] FROM EXTERNAL PROVIDER;
GO
ALTER ROLE db_datareader ADD MEMBER [schema-operator-tester];
ALTER ROLE db_datawriter ADD MEMBER [schema-operator-tester];
ALTER ROLE db_owner ADD MEMBER [schema-operator-tester];
GRANT EXECUTE TO [schema-operator-tester];
"""

echo "Done"