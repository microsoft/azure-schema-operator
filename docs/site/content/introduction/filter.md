# Target Filter

The operator uses the `TargetFilter` struct to Filter and aquire execution targets.
The semantics change a bit between the different db technologies to match common use patterns.

Although the general notions stay the same.

`ClusterUris` holds a list of clusters/servers/Eventhub namespaces.
`Create` flag indicates if we should create the DB/schema/registry if missing.
`Regexp` flag indicates if we should regard the filter values as regular expressions or exact match.

## Kusto filtering

In Kusto a common multi-tenantcy solution is DB per tenant,
To support this scenario we use the `DB` field as a regular expression to filter all the database names in the clusters.

Sometimes an external system is used to determain the schema type instead of DB name, e.g. if we have different tier users.
To support this scenario we have a `Webhook` & `Label` system, we will make a rest call to that webhook and passing the label.
The response is expected to be a json array with database names on which we should apply the schema.

## SQL Server filtering

In Sql Server a common multi-tenantcy solution is Schema per tenant,
to support this scenario we use the `Schema` field as a regular expression to filter all the schema names.
**Note** In this scenario we regard the DB name as exact match.

## Eventhub schema registry

In Eventhubs we can only define one schema registry - we match the name according to the `DB` field.
