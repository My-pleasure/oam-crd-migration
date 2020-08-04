# oam-crd-migration
A tool to help you migrate OAM CRDs from v1alpha1 to v1alpha2

# To do
- [x] [crd conversion webhook](https://github.com/kubernetes/kubernetes/tree/master/test/images/agnhost)
- [ ] [storage version migrator](https://github.com/kubernetes-sigs/kube-storage-version-migrator)

# User guide for examples
## Pre-requisites
- Clusters with old versions of CRD
```
kubectl kustomize ./crd/bases/ | kubectl apply -f -

kubectl apply -f example/example-instance.yaml
```
## The conversion process
- Build image
```
docker build -t example:v0.1 .
```
- Deploy a deployment and a service for webhook
```
kubectl apply -f deploy/webhook.yaml
```
- Patch new versions and conversion strategy to CRD
```
kubectl kustomize ./crd/patches | kubectl apply -f -
```
- Verify that the old and new version objects are available
```
# kubectl describe ex example-test

Name:         example-test
Namespace:    default
Labels:       <none>
Annotations:  kubectl.kubernetes.io/last-applied-configuration:
                {"apiVersion":"core.oam.dev/v1alpha1","hostPort":"localhost:1234","kind":"Example","metadata":{"annotations":{},"name":"example-test","nam...
API Version:  core.oam.dev/v1alpha2
Host:         localhost
Kind:         Example
...
Port:                1234
Events:              <none>

# kubectl describe ex.v1alpha1.core.oam.dev example-test

Name:         example-test
Namespace:    default
Labels:       <none>
Annotations:  kubectl.kubernetes.io/last-applied-configuration:
                {"apiVersion":"core.oam.dev/v1alpha1","hostPort":"localhost:1234","kind":"Example","metadata":{"annotations":{},"name":"example-test","nam...
API Version:  core.oam.dev/v1alpha1
Host Port:    localhost:1234
Kind:         Example
```


# The migration process (simple)
1. Generate certificate and secret. And deploy crd conversion webhook.
2. Use Kustomize to add new crd version and conversion strategy to old crd version. And set `storage` to `true` for the new version.
3. Migrate stored objects to the new version.
4. Ensure all clients are fully migrated to the new version and all stored objects are new version. Then set `served` to `false` for the old version.
5. Remove the old version from crd and drop the conversion support.

More details see [this](https://github.com/crossplane/oam-kubernetes-runtime/issues/108).