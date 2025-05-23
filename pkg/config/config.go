package config

import (
	"path/filepath"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog/v2"
)

// GetKubeConfig returns a Kubernetes client configuration
func GetKubeConfig() (*rest.Config, error) {
	var kubeconfig string

	if home := homedir.HomeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		klog.Warning("External kubeconfig not found or invalid, falling back to in-cluster config")
		return rest.InClusterConfig()
	}

	return config, nil
}

