# Helm resource watcher Quick start
Here is an example of how to 

## Download the github repo
```shell
git clone 
```

## Usage example
1. Install deployment based workloadDefinition
```shell
make install
make run
```

2. Create a Ryan object, here is an example 
```yaml
apiVersion: practice.shipa.io/v1alpha1
kind: Ryan
metadata:
  name: helm-watcher
spec:
  ResourceName: ryan-test
```

3. Apply the object
```shell
kubectl apply -f ryan.yaml
```

4. Watch the object status
```shell
kubectl apply -f docs/examples/deployment-rollout/app-rollout-pause.yaml
```
Check the status of the ApplicationRollout and see the step by step rolling out.
This rollout will pause after the second batch.

6. Apply the application deployment that completes the rollout
```shell
kubectl apply -f docs/examples/deployment-rollout/app-rollout-finish.yaml
```
Check the status of the AppRollout and see the rollout completes, and the
AppRollout's "Rolling State" becomes `rolloutSucceed`