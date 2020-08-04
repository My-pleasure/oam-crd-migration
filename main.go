package main

import (
	"crypto/tls"
	"fmt"
	"github.com/spf13/cobra"
	"net/http"

	config2 "sigs.k8s.io/controller-runtime/pkg/client/config"

	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"

	"github.com/My-pleasure/oam-crd-migration/converter"
)

var (
	certFile string
	keyFile  string
	port     int
)

var ConversionWebhookArgs = &cobra.Command{
	Use:  "crd-conversion-webhook",
	Args: cobra.MaximumNArgs(0),
	Run:  transferArgs,
}

func init() {
	ConversionWebhookArgs.Flags().StringVar(&certFile, "tls-cert-file", "",
		"File containing the default x509 Certificate for HTTPS. (CA cert, if any, concatenated "+
			"after server cert.")
	ConversionWebhookArgs.Flags().StringVar(&keyFile, "tls-private-key-file", "",
		"File containing the default x509 private key matching --tls-cert-file.")
	ConversionWebhookArgs.Flags().IntVar(&port, "port", 443,
		"Secure port that the webhook listens on")
}

// Config contains the server (the webhook) cert and key.
type Config struct {
	CertFile string
	KeyFile  string
}

func main() {
	ConversionWebhookArgs.Execute()
}

func transferArgs(cmd *cobra.Command, args []string) {
	config := Config{CertFile: certFile, KeyFile: keyFile}

	http.HandleFunc("/exampleconvert", converter.ServeExampleConvert)
	http.HandleFunc("/appconfigconvert", converter.ServeAppConfigConvert)
	clientset := getClient()
	server := &http.Server{
		Addr:      fmt.Sprintf(":%d", port),
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
