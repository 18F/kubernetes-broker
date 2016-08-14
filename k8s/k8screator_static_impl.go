package k8s

import (
	"errors"
)

type K8sStaticConnector struct {
	creds K8sClusterCredentials
}

func NewK8sStaticConnector(creds K8sClusterCredentials) *K8sStaticConnector {
	return &K8sStaticConnector{creds: creds}
}

func (k *K8sStaticConnector) DeleteCluster(org string) error {
	return errors.New("Delete operation not supported")
}

func (k *K8sStaticConnector) GetCluster(org string) (int, K8sClusterCredentials, error) {
	return 0, k.creds, nil
}

func (k *K8sStaticConnector) GetOrCreateCluster(org string) (K8sClusterCredentials, error) {
	_, creds, err := k.GetCluster(org)
	return creds, err
}

func (k *K8sStaticConnector) PostCluster(org string) (int, error) {
	return 0, errors.New("Post operation not supported")
}

func (k *K8sStaticConnector) GetClusters() ([]K8sClusterCredentials, error) {
	_, cluster, _ := k.GetCluster("")
	return []K8sClusterCredentials{cluster}, nil
}
