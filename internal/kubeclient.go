package internal

import (
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"
)

type KubeClient struct {
	*kubernetes.Clientset
}

// NewKubeClient(
func NewKubeClient(kubecfg *KubeConfig) (*KubeClient, error) {
	if kubecfg == nil {
		return nil, ErrInvalidParams
	}

	raw, err := kubecfg.RawBytes()
	if err != nil {
		return nil, err
	}

	clientCfg, err := clientcmd.RESTConfigFromKubeConfig(raw)
	if err != nil {
		return nil, err
	}

	clusterSet, err := kubernetes.NewForConfig(clientCfg)
	if err != nil {
		return nil, err
	}

	return &KubeClient{clusterSet}, nil
}
