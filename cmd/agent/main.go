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
		fmt.Println("[Erro] A variável APP_LABEL é obrigatória. Ex: app=minha-app")
		os.Exit(1)
	}

	col := collector.NewCollector()
	col.Start() // Inicia os cronômetros (loopCPU e loopPrinter)

	watcher, err := discovery.NewPodWatcher(col)
	if err != nil {
		fmt.Printf("[Erro Crítico] Falha ao iniciar o K8s Watcher: %v\n", err)
		os.Exit(1)
	}
	watcher.Start() // Começa a ouvir os eventos

	// Trava a thread principal para o programa rodar para sempre
	select {}
}
