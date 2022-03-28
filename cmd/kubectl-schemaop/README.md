# kubectl-schemaop

A kubectl plugin to manage schema rollouts.  
This is greatly influenced by the rollout kubectl sub command.  

## sample runs

Current schema rollout status:

```bash
$ kubectl schemaop status --namespace default --name master-test-template              
  NAMESPACE  NAME                    REVISION  EXECUTED  FAILED  RUNNING  SUCCEEDED  
  default    master-test-template-1  1         true      0       0        1          
```

to list the schema changes history use:

```bash
$ kubectl schemaop history --namespace default --name master-test-template
  NAMESPACE  NAME                    REVISION  
  default    master-test-template-0  0         
  default    master-test-template-1  1         
```

To see specific revision details add the `--revision=<num>` flag:

```bash
$ kubectl schemaop history --namespace default --name master-test-template --revision=1
  NAMESPACE  NAME                    REVISION  EXECUTED  FAILED  RUNNING  SUCCEEDED  
  default    master-test-template-1  1         true      0       0        1          
```

## Development

Build the plugin with `make kubectl-schemaop` , the resulting binary will be generated in the `bin` folder.  
(add `PATH=$PATH:./bin` to test as the plugin needs to be in `$PATH` )  
