package kubetool

import (
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// Kubetool permit to connect on Kubernetes cluster
type Kubetool struct {
	client  kubernetes.Interface
	dclient dynamic.Interface
}

// NewConnexion permit to connect on Kubernetes cluster from config file
func NewConnexion(configPath string) (cmd *Kubetool, err error) {
	config, err := clientcmd.BuildConfigFromFlags("", configPath)
	if err != nil {
		return cmd, err
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return cmd, err
	}

	dclient, err := dynamic.NewForConfig(config)
	if err != nil {
		return cmd, err
	}

	cmd = &Kubetool{
		client:  client,
		dclient: dclient,
	}

	return cmd, err
}

func NewConnexionFromClient(client kubernetes.Interface, dClient dynamic.Interface) (cmd *Kubetool) {
	return &Kubetool{
		client:  client,
		dclient: dClient,
	}
}
