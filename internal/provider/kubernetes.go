package provider

import (
	"context"
	"fmt"
	"os"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type K3sProvider struct {
	client kubernetes.Interface
}

func NewK3sProvider() (*K3sProvider, error) {
	var cfg *rest.Config
	var err error

	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig != "" {
		cfg, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		cfg, err = rest.InClusterConfig()
	}
	if err != nil {
		return nil, fmt.Errorf("k3s config: %w", err)
	}

	cfg.Insecure = true
	cfg.CAData = nil
	cfg.CAFile = ""

	c, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("k3s client: %w", err)
	}

	return &K3sProvider{client: c}, nil
}

func (p *K3sProvider) Name() string { return "k3s" }

func (p *K3sProvider) CreateGroup(ctx context.Context, groupName string) error {
	_, err := p.client.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: groupName},
	}, metav1.CreateOptions{})
	if apierrors.IsAlreadyExists(err) {
		return nil
	}
	return err
}

func (p *K3sProvider) DeleteGroup(ctx context.Context, groupName string) error {
	err := p.client.CoreV1().Namespaces().Delete(ctx, groupName, metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		return nil
	}
	return err
}
