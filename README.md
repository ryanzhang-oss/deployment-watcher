# Helm resource watcher Quick start
This is a toy controller built by kubebuilder that watches a k8s resource managed by Helm. The 
controller will keep track of the helm application version after it locates the resource. 

It only supports deployment type of resource for now.

## Download the github repo
```shell
git clone git@github.com:ryanzhang-oss/deployment-watcher.git
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

3. Install an application through helm
```shell
helm repo add podinfo https://stefanprodan.github.io/podinfo
helm repo update
helm upgrade --install helmapp podinfo/podinfo --version 5.0.0 --wait
```

4. Apply the watcher object
```shell
kubectl apply -f doc/examples/podinfo.yaml
```

5. Watch the watcher object
```shell
 kubectl get Ryan --watch
 
 NAME           APPVERSION
helm-watcher   5.0.0
helm-watcher   5.0.3
helm-watcher   5.1.1
 ```

6. Upgrade an application through helm
```shell
helm upgrade --install helmapp podinfo/podinfo --version 5.0.3 --wait
helm upgrade --install helmapp podinfo/podinfo --version 5.1.1 --wait
```