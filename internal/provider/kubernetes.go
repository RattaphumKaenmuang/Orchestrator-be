package provider

import (
	"context"
	"fmt"
	"os"

	"orchestrator/internal/model"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
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

func (p *K3sProvider) CreateGroup(ctx context.Context, groupName string) (string, error) {
	ns, err := p.client.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: groupName},
	}, metav1.CreateOptions{})

	if apierrors.IsAlreadyExists(err) {
		return groupName, nil
	}
	if err != nil {
		return "", err
	}

	return ns.Name, nil
}

func (p *K3sProvider) DeleteGroup(ctx context.Context, groupName string) error {
	err := p.client.CoreV1().Namespaces().Delete(ctx, groupName, metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		return nil
	}
	return err
}

func (p *K3sProvider) GroupExists(ctx context.Context, groupName string) (bool, error) {
	_, err := p.client.CoreV1().Namespaces().Get(ctx, groupName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (p *K3sProvider) CreateInstance(ctx context.Context, group *model.Group, instance *model.Instance) (string, error) {
	resources := corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(fmt.Sprintf("%d", instance.CPU)),
			corev1.ResourceMemory: resource.MustParse(fmt.Sprintf("%dMi", instance.RAM)),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(fmt.Sprintf("%d", instance.CPU)),
			corev1.ResourceMemory: resource.MustParse(fmt.Sprintf("%dMi", instance.RAM)),
		},
	}

	// if instance.GPU > 0 {
	// 	gpuQty := resource.MustParse(fmt.Sprintf("%d", instance.GPU))
	// 	resources.Requests["nvidia.com/gpu"] = gpuQty
	// 	resources.Limits["nvidia.com/gpu"] = gpuQty
	// }

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: group.K3sNamespace,
			Labels: map[string]string{
				"orchestrator/group":    group.Name,
				"orchestrator/instance": instance.Name,
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:      "main",
					Image:     "cirros",
					Command:   []string{"sleep", "infinity"},
					Resources: resources,
				},
			},
			RestartPolicy: corev1.RestartPolicyAlways,
		},
	}

	created, err := p.client.CoreV1().Pods(group.K3sNamespace).Create(ctx, pod, metav1.CreateOptions{})
	if err != nil {
		return "", fmt.Errorf("k3s create pod: %w", err)
	}

	return string(created.UID), nil
}
