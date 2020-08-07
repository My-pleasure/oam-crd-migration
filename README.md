# oam-crd-migration
A tool to help you migrate OAM CRDs from v1alpha1 to v1alpha2.

More details see [this](https://github.com/crossplane/oam-kubernetes-runtime/issues/108).

# To do
- [x] [crd conversion webhook](https://github.com/kubernetes/kubernetes/tree/master/test/images/agnhost)
- [x] [storage version migrator](https://github.com/kubernetes-sigs/kube-storage-version-migrator)
- [x] [a golang script](https://github.com/elastic/cloud-on-k8s/issues/2196) to remove old versions from CRD `status.storedVersions`

# User guide for appconfig examples
This guide is for appconfig CRD version migration proess, but it is not complete. You can refer to the [document](./sample/README.md) for another demo.
## Pre-requisites
- Clusters with old versions of CRD
    ```
    kubectl kustomize ./crd/bases/ | kubectl apply -f -
    
    kubectl apply -f crd/appconfig_v1alpha1_example.yaml
    ```
## The conversion process
- Create secret for ssl certificates
    ```
    curl -sfL https://raw.githubusercontent.com/crossplane/oam-kubernetes-runtime/master/hack/ssl/ssl.sh | bash -s crd-conversion-webhook default
    
    kubectl create secret generic webhook-server-cert --from-file=tls.key=./crd-conversion-webhook.key --from-file=tls.crt=./crd-conversion-webhook.pem
    ```
- Create CA Bundle info and inject into the CRD definition
    ```
    caValue=`kubectl config view --raw --minify --flatten -o jsonpath='{.clusters[].cluster.certificate-authority-data}'`
    
    sed -i 's/${CA_BUNDLE}/'"$caValue"'/g' ./crd/patches/crd_conversion_examples.yaml
    ```
- Build image and deploy a deployment and a service for webhook
    ```
    docker build -t example:v0.1 .

    kubectl apply -f deploy/webhook.yaml
    ```
- Patch new versions and conversion strategy to CRD
    ```
    kubectl get crd applicationconfigurations.core.oam.dev -o yaml >> ./crd/patches/temp.yaml
  
    kubectl kustomize ./crd/patches | kubectl apply -f -
    ```
- Verify that the old and new version objects are available (only APIVersion has been converted)
    ```
    # kubectl describe applicationconfigurations complete-app
    
    Name:         complete-app
    Namespace:    default
    Labels:       <none>
    Annotations:  API Version:  core.oam.dev/v1alpha2
    Kind:         ApplicationConfiguration
    ...
    
    # kubectl describe applicationconfigurations.v1alpha1.core.oam.dev complete-app
    
    Name:         complete-app
    Namespace:    default
    Labels:       <none>
    Annotations:  API Version:  core.oam.dev/v1alpha1
    Kind:         ApplicationConfiguration
    ...
    ```