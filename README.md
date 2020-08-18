# oam-crd-migration
A tool to help you migrate OAM CRDs from v1alpha1 to v1alpha2.

More details see [this](https://github.com/crossplane/oam-kubernetes-runtime/issues/108).

# To do
- [x] [crd conversion webhook](https://github.com/kubernetes/kubernetes/tree/master/test/images/agnhost)
- [x] [storage version migrator](https://github.com/kubernetes-sigs/kube-storage-version-migrator)
- [x] [a golang script](https://github.com/elastic/cloud-on-k8s/issues/2196) to remove old versions from CRD `status.storedVersions`

# For developers
- [converter/framework.go:](converter/framework.go)
定义了如何处理 ConversionReview 的请求与响应（Request、Response），一般不需要更改。
- [converter/converter.go:](converter/converter.go)
定义了针对 ApplicationConfigurations 的转换函数，此处的输入输出是 unstructured 结构的 ApplicationConfigurations，若要更改输入输入输出，则需更改 framework.go 中如下两处：
    ```
    #L24
    type convertFunc func(Object *unstructured.Unstructured, version string) (*unstructured.Unstructured, metav1.Status)
    #L74
    func doConversionV1(convertRequest *v1.ConversionRequest, convert convertFunc) *v1.ConversionResponse {
        var convertedObjects []runtime.RawExtension
        for _, obj := range convertRequest.Objects {
            cr := unstructured.Unstructured{}
            if err := cr.UnmarshalJSON(obj.Raw); err != nil {
                klog.Error(err)
                return &v1beta1.ConversionResponse{
                    Result: metav1.Status{
                        Message: fmt.Sprintf("failed to unmarshall object (%v) with error: %v", string(obj.Raw), err),
                        Status:  metav1.StatusFailure,
                    },
                }
            }
        ...
    }
    ```
- [converter/plugin.go:](converter/plugin.go)
添加和修改具体转换逻辑都在 plugin.go 文件中。
此文件中定义了针对 ApplicationConfigurations 中 components 和 traits 字段转换的接口和实现的具体逻辑：
    ```
    type Converter interface {
        ConvertComponent(v1alpha1Component) (v1alpha2Component, v1alpha2.Component, error)
        ConvertTrait(v1alpha1Trait) (v1alpha2Trait, error)
    }
    ```

# User guide for appconfig examples
## Pre-requisites
- Clusters with old versions of CRD
    ```
    kubectl kustomize ./crd/bases/ | kubectl apply -f -
    
    kubectl apply -f crd/appconfig_v1alpha1_example.yaml
    ```
- Because webhook is deployed in default namespace in this demo, permissions should be assigned.
    ```
    kubectl apply -f crd/role-binding.yaml
    ```
## The conversion process
- Create secret for ssl certificates
    ```
    curl -sfL https://raw.githubusercontent.com/crossplane/oam-kubernetes-runtime/master/hack/ssl/ssl.sh | bash -s oam-crd-conversion default
    
    kubectl create secret generic webhook-server-cert --from-file=tls.key=./oam-crd-conversion.key --from-file=tls.crt=./oam-crd-conversion.pem
    ```
- Create CA Bundle info and inject into the CRD definition
    ```
    caValue=`kubectl config view --raw --minify --flatten -o jsonpath='{.clusters[].cluster.certificate-authority-data}'`
    
    sed -i 's/${CA_BUNDLE}/'"$caValue"'/g' ./crd/patches/crd_conversion_applicationconfigurations.yaml
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
- Verify that the old and new version objects are available
    ```
    # kubectl describe applicationconfigurations complete-app
    
    Name:         complete-app
    Namespace:    default
    Labels:       <none>
    Annotations:  API Version:  core.oam.dev/v1alpha2
    Kind:         ApplicationConfiguration
    ...
      Traits:
        Trait:
          API Version:  core.oam.dev/v1alpha2
          Kind:         RollOutTrait
          Metadata:
            Name:  rollout
          Spec:
            Auto:               true
            Batch Interval:     5
            Batches:            2
            Canary Replicas:    0
            Instance Interval:  1
    
    # kubectl describe applicationconfigurations.v1alpha1.core.oam.dev complete-app
    
    Name:         complete-app
    Namespace:    default
    Labels:       <none>
    Annotations:  API Version:  core.oam.dev/v1alpha1
    Kind:         ApplicationConfiguration
    ...
      Traits:
        Name:  rollout
        Properties:
          Name:   canaryReplicas
          Value:  0
          Name:   batches
          Value:  2
          Name:   batchInterval
          Value:  5
          Name:   instanceInterval
          Value:  1
          Name:   auto
          Value:  true
    ```
## Update existing objects
Here we use kube-storage-version-migrator as an example, you can write Go scripts instead.
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
  
    NAME                       STATUS
    ...                        ...
    applicationconfigurations  SUCCEEDED
    ...                        ...
    ```
## Remove old versions
- Run the golang script that removes old versions from CRD `status.storedVersions` field
    ```
    go run remove/remove.go
  
    updated applicationconfigurations.core.oam.dev CRD status storedVersions: [v1alpha2]
    ```
- Verify the script runs successfully
    ```
    kubectl describe crd applicationconfigurations.core.oam.dev
  
    Name:         applicationconfigurations.core.oam.dev
    Namespace:    
    ...
      Stored Versions:
        v1alpha2
    Events:  <none>
    ```
- Remove the old version from the CustomResourceDefinition spec.versions list
    ```
    kubectl get crd applicationconfigurations.core.oam.dev -o yaml >> ./crd/complete/temp.yaml
  
    kubectl kustomize ./crd/complete | kubectl apply -f -
    ```