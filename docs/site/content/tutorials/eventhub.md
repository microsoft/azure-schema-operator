# Eventhub schema regitery tutorial

Creating the ConfigMap:

```bash
kubectl create configmap event-demo --from-literal templateName="schemaop"  --from-literal group="testsgr" \
--from-file=schema=docs/samples/eventhubs/avro-schema.json   
```

next we need to define a `SchemaDeployment` object that will reference the `ConfigMap`.

```yaml
apiVersion: dbschema.microsoft.com/v1alpha1
kind: SchemaDeployment
metadata:
  name: eventhub-schema-demo
spec:
  type: eventhub
  applyTo:
    clusterUris: ['schematest.servicebus.windows.net']
  failIfDataLoss: false
  failurePolicy: abort
  source:
    name: event-demo
    namespace: default
```

and apply it via kubectl:

```bash
kubectl apply -f docs/samples/eventhubs/eventhub-schema-demo.yaml
```

To demonstrate schema evolution, we will add a new field:

```json
{
  "name": "description",
  "type": "string"
}
```

let's run:

```bash
kubectl create configmap event-demo --from-literal templateName="schemaop"  --from-literal group="testsgr" \
--from-file=schema=docs/samples/eventhubs/avro-schema-v2.json --dry-run=client -o yaml | kubectl apply -f -
```

To see the deplopyment status we can check the history:

```bash
kubectl schemaop history --namespace default --name eventhub-schema-demo
  NAMESPACE  NAME                    REVISION  
  default    eventhub-schema-demo-0  0         
  default    eventhub-schema-demo-1  1        
```
