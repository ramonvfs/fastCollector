package discovery

import (
	"context"
	"fmt"
	"os"

	"fastcollector/internal/collector"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type PodWatcher struct {
	clientset *kubernetes.Clientset
	collector *collector.Collector
}

func NewPodWatcher(col *collector.Collector) (*PodWatcher, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("In-Cluster config error: %v", err)
	}

	cs, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("Error creating clientset: %v", err)
	}

	return &PodWatcher{clientset: cs, collector: col}, nil
}

func (w *PodWatcher) Start() {
	appLabel := os.Getenv("APP_LABEL")
	nodeName := os.Getenv("NODE_NAME")

	fmt.Printf("[Watcher] Iniciando busca por pods com label '%s' no nó '%s'...\n", appLabel, nodeName)

	opts := metav1.ListOptions{LabelSelector: appLabel}
	watcher, err := w.clientset.CoreV1().Pods("").Watch(context.Background(), opts)

	if err != nil {
		fmt.Printf("[Error] Failed to start pod observation: %v\n", err)
		return
	}

	// Goroutine that stays indefinitely reading events from Kubernetes
	go func() {
		for event := range watcher.ResultChan() {
			pod, ok := event.Object.(*corev1.Pod)
			if !ok {
				continue
			}

			if nodeName != "" && pod.Spec.NodeName != nodeName {
				continue
			}

			switch event.Type {
			case watch.Added:
				podUID := string(pod.UID)

				cpuPath, err := ResolveCPUPath(podUID)
				if err != nil {
					fmt.Printf("[Warning] %v\n", err)
					continue
				}

				netPath, err := ResolveNetPath(podUID)
				if err != nil {
					fmt.Printf("[Warning] %v\n", err)
					continue
				}

				w.collector.AddPod(pod.Name, cpuPath, netPath)

			case watch.Deleted:
				w.collector.RemovePod(pod.Name)
			}
		}
	}()
}
