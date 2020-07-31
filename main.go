package main

import (
	"crypto/tls"
	"net/http"

	config2 "sigs.k8s.io/controller-runtime/pkg/client/config"

	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"

	"github.com/My-pleasure/oam-crd-migration/converter"
)

// TODO：Allow port, cert and other information to be passed in as parameters
//var ConversionWebhookArgs = &cobra.Command{
//	Use: "crd-conversion-webhook",
//	Args: cobra.MaximumNArgs(0),
//	Run: main,
//}

// Config contains the server (the webhook) cert and key.
type Config struct {
	CertFile string
	KeyFile  string
}

func main() {
	// These default values are temporarily
	config := Config{CertFile: "/etc/webhook/cert/tls.crt", KeyFile: "/etc/webhook/cert/tls.key"}

	http.HandleFunc("/exampleconvert", converter.ServeExampleConvert)
	http.HandleFunc("/readyz", func(w http.ResponseWriter, req *http.Request) { w.Write([]byte("ok")) })
	clientset := getClient()
	server := &http.Server{
		Addr:      ":9443",
		TLSConfig: configTLS(config, clientset),
	}
	err := server.ListenAndServeTLS("", "")
	if err != nil {
		panic(err)
	}
}

// Get a clientset with in-cluster config.
func getClient() *kubernetes.Clientset {
	config, err := config2.GetConfig()
	if err != nil {
		klog.Fatal(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Fatal(err)
	}
	return clientset
}

func configTLS(config Config, clientset *kubernetes.Clientset) *tls.Config {
	sCert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
	if err != nil {
		klog.Fatal(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{sCert},
		// TODO: uses mutual tls after we agree on what cert the apiserver should use.
		// ClientAuth:   tls.RequireAndVerifyClientCert,
	}
}
