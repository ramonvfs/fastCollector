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

// NewPodWatcher cria o cliente de comunicação com a API do K8s (in-cluster).
func NewPodWatcher(col *collector.Collector) (*PodWatcher, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("erro na config in-cluster (você está rodando fora do pod?): %v", err)
	}

	cs, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar clientset: %v", err)
	}

	return &PodWatcher{clientset: cs, collector: col}, nil
}

// Start inicia o "Watch" na API do Kubernetes.
func (w *PodWatcher) Start() {
	appLabel := os.Getenv("APP_LABEL")
	nodeName := os.Getenv("NODE_NAME")

	fmt.Printf("[Watcher] Iniciando busca por pods com label '%s' no nó '%s'...\n", appLabel, nodeName)

	opts := metav1.ListOptions{LabelSelector: appLabel}
	watcher, err := w.clientset.CoreV1().Pods("").Watch(context.Background(), opts)

	if err != nil {
		fmt.Printf("[Erro] Falha ao iniciar observação de pods: %v\n", err)
		return
	}

	// Goroutine que fica infinitamente lendo os eventos do Kubernetes
	go func() {
		for event := range watcher.ResultChan() {
			pod, ok := event.Object.(*corev1.Pod)
			if !ok {
				continue
			}

			// Se NODE_NAME estiver definido, ignoramos os pods de outros nós.
			if nodeName != "" && pod.Spec.NodeName != nodeName {
				continue
			}

			switch event.Type {
			case watch.Added:
				// Pegamos o UID do Pod recém descoberto
				podUID := string(pod.UID)

				// Chamamos a sua função para descobrir o caminho
				cpuPath, err := ResolveCPUPath(podUID)
				if err != nil {
					fmt.Printf("[Aviso] %v\n", err)
					continue
				}

				// Adicionamos ao coletor
				w.collector.AddPod(pod.Name, cpuPath)

			case watch.Deleted:
				// Se o pod foi apagado, mandamos o coletor remover
				w.collector.RemovePod(pod.Name)
			}
		}
	}()
}
