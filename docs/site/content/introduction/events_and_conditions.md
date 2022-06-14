# Events and Conditions

Azure Schema Operator exposes varios events and conditions to allow for simpler status observation.

## Conditions

We can wait for a `SchemaDeployment` to finish by waiting on the `Execution` condition:

```bash
➜ kubectl wait --for=condition=Execution --timeout=10s   schemadeployment/master-test-template
schemadeployment.dbschema.microsoft.com/master-test-template condition met
```

The condition is also displayed when we get the `SchemaDeployment` object:

```bash
➜ kubectl get schemadeployments master-test-template
NAME                   TYPE    EXECUTED
master-test-template   kusto   True
```

In case of failuer the `Execution` condition will be marked as such:

```bash
➜ kubectl get schemadeployments master-test-template
NAME                   TYPE    EXECUTED
master-test-template   kusto   False
```

## Events

Dureing the deployment process events will be reported on the different steps and changes that occur.

Events differ according to the type and required deployment, and can easily be seen from `kubectl`:

```bash {linenos=table}
➜ kubectl get events --field-selector involvedObject.name=master-test-template
LAST SEEN   TYPE     REASON     OBJECT                                  MESSAGE
6m36s       Normal   Created    schemadeployment/master-test-template   Created versioned deployment "master-test-template-0"
5m36s       Normal   Executed   schemadeployment/master-test-template   Scheme was deployed
```
