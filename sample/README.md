# User guide for examples
## Pre-requisites
- Clusters with old versions of CRD
    ```
    kubectl kustomize ./sample/bases/ | kubectl apply -f -
    
    kubectl apply -f sample/example-instance.yaml
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
    kubectl kustomize ./sample/patches | kubectl apply -f -
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
## Update existing objects
- Run the storage Version migrator
    ```
    git clone https://github.com/kubernetes-sigs/kube-storage-version-migrator
  
    sed -i 's/kube-system/default/g' ./Makefile
  
    make local-manifests
  
    sed -i '1,5d' ./manifests.local/namespace-rbac.yaml
  
    pushd manifests.local && kubectl apply -k ./ && popd
    ```
- Verify the migration is "SUCCEEDED"
    ```
    kubectl get storageversionmigrations -o=custom-columns=NAME:.spec.resource.resource,STATUS:.status.conditions[0].type
  
    NAME                     STATUS
    ...                      ...
    examples                 Succeeded
    ...                      ...
    ```
## Remove old versions
- Run the golang script that removes old versions from CRD `status.storedVersions` field
    ```
    go run remove/remove.go
  
    updated examples.core.oam.dev CRD status storedVersions: [v1alpha2]
    ```
- Verify the script runs successfully
    ```
    kubectl describe crd examples.core.oam.dev
  
    Name:         examples.core.oam.dev
    Namespace:    
    ...
      Stored Versions:
        v1alpha2
    Events:  <none>
    ```
- Remove the old version from the CustomResourceDefinition spec.versions list
    ```
    kubectl kustomize ./sample/complete | kubectl apply -f -
    ```