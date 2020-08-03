# oam-crd-migration
A tool to help you migrate OAM CRDs from v1alpha1 to v1alpha2

# To do
- [ ] [crd conversion webhook](https://github.com/kubernetes/kubernetes/tree/master/test/images/agnhost)
- [ ] [storage version migrator](https://github.com/kubernetes-sigs/kube-storage-version-migrator)

# User guide


# The migration process (simple)
1. Generate certificate and secret. And deploy crd conversion webhook.
2. Use Kustomize to add new crd version and conversion strategy to old crd version. And set `storage` to `true` for the new version.
3. Migrate stored objects to the new version.
4. Ensure all clients are fully migrated to the new version and all stored objects are new version. Then set `served` to `false` for the old version.
5. Remove the old version from crd and drop the conversion support.

More details see [this](https://github.com/crossplane/oam-kubernetes-runtime/issues/108).