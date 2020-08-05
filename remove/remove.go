package main

import (
	"context"
	"fmt"
	"os"

	apiextension "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func exitOnErr(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	//  create k8s client
	ctx := context.Background()
	cfg, err := config.GetConfig()
	client, err := apiextension.NewForConfig(cfg)
	exitOnErr(err)

	updateStatus(client, ctx, "examples.core.oam.dev")
}

// updateStatus remove v1alpha1 from CRD status
func updateStatus(client *apiextension.Clientset, ctx context.Context, gvk string) {
	// retrieve CRD
	crd, err := client.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, gvk, v1.GetOptions{})
	exitOnErr(err)
	// remove v1alpha1 from its status
	oldStoredVersions := crd.Status.StoredVersions
	newStoredVersions := make([]string, 0, len(oldStoredVersions))
	for _, stored := range oldStoredVersions {
		if stored != "v1alpha1" {
			newStoredVersions = append(newStoredVersions, stored)
		}
	}
	crd.Status.StoredVersions = newStoredVersions
	// update the status sub-resource
	crd, err = client.ApiextensionsV1().CustomResourceDefinitions().UpdateStatus(ctx, crd, v1.UpdateOptions{})
	exitOnErr(err)
	fmt.Println("updated", gvk, "CRD status storedVersions:", crd.Status.StoredVersions)
}
