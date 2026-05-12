package main

import (
	"fastcollector/internal/collector"
	"fastcollector/internal/discovery"
	"fmt"
	"os"
)

func main() {
	fmt.Println("=== Iniciando DaemonSet de Monitoramento de CPU ===")

	// Verificações de segurança para as variáveis de ambiente
	if os.Getenv("APP_LABEL") == "" {
		fmt.Println("[Error] A variável APP_LABEL é obrigatória. Ex: app=minha-app")
		os.Exit(1)
	}

	col, err := collector.NewCollector("/var/log/ia-data/cpu_dataset.csv")
	if err != nil {
		fmt.Printf("[Error] Falha ao criar o coletor: %v\n", err)
		os.Exit(1)
	}
	col.Start() //  (loopCPU and loopPrinter)

	watcher, err := discovery.NewPodWatcher(col)
	if err != nil {
		fmt.Printf("[Error] Falha ao iniciar o K8s Watcher: %v\n", err)
		os.Exit(1)
	}
	watcher.Start()

	select {}
}
