package collector

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// MetricPoint guarda o valor calculado no nosso ciclo de 10ms.
type MetricPoint struct {
	Value float64
}

// PodTarget mantém o arquivo real aberto para cada Pod
type PodTarget struct {
	ID           string
	CPUFile      *os.File
	LastCPUUsage uint64
}

type Collector struct {
	mu        sync.RWMutex
	targets   map[string]*PodTarget
	CPUBuffer []MetricPoint
}

func NewCollector() *Collector {
	return &Collector{
		targets:   make(map[string]*PodTarget),
		CPUBuffer: make([]MetricPoint, 0),
	}
}

func (c *Collector) AddPod(podID, cpuPath string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	fCPU, err := os.Open(cpuPath)
	if err != nil {
		fmt.Printf("[Erro] Não foi possível abrir %s para o Pod %s: %v\n", cpuPath, podID, err)
		return
	}

	c.targets[podID] = &PodTarget{
		ID:      podID,
		CPUFile: fCPU,
	}
	fmt.Printf("[Sucesso] Monitorando CPU do Pod: %s\n", podID)
}

func (c *Collector) RemovePod(podID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if t, ok := c.targets[podID]; ok {
		t.CPUFile.Close() // Muito importante para evitar memory leak
		delete(c.targets, podID)
		fmt.Printf("[Info] Pod removido do monitoramento: %s\n", podID)
	}
}

func (c *Collector) Start() {
	go c.loopCPU()
	go c.loopPrinter()
}

func (c *Collector) loopCPU() {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for range ticker.C {
		c.collectCPU()
	}
}

// loopPrinter exibe o resultado agrupado no terminal a cada 1 segundo.
func (c *Collector) loopPrinter() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		c.mu.Lock()
		if len(c.CPUBuffer) > 0 {
			// Pega o último cálculo do buffer
			last := c.CPUBuffer[len(c.CPUBuffer)-1]
			fmt.Printf("[MÉTRICA 1s] CPU Média do Nó: %.2f us | Amostras no segundo: %d\n", last.Value, len(c.CPUBuffer))
			c.CPUBuffer = nil // Limpa o buffer para o próximo segundo
		}
		c.mu.Unlock()
	}
}
