package collector

import (
	"fmt"
	"os"
	"sync"
	"time"
)

type MetricPoint struct {
	Timestamp int64
	Value     float64
}

type PodTarget struct {
	ID           string
	CPUFile      *os.File
	LastCPUUsage uint64
	NetPID       string
	LastNetRx    uint64
}

type Collector struct {
	mu      sync.RWMutex
	targets map[string]*PodTarget
	LogFile *os.File
}

func NewCollector(logPath string) (*Collector, error) {
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	info, _ := f.Stat()
	if info.Size() == 0 {
		// Write header with timestamp,cpu_millicore and num_rx
		f.WriteString("timestamp,cpu_millicore\n")
	}

	return &Collector{
		targets: make(map[string]*PodTarget),
		LogFile: f,
	}, nil
}

func (c *Collector) AddPod(podID, cpuPath string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	fCPU, err := os.Open(cpuPath)
	if err != nil {
		fmt.Printf("[Error] Failed to open %s for Pod %s: %v\n", cpuPath, podID, err)
		return
	}

	c.targets[podID] = &PodTarget{
		ID:      podID,
		CPUFile: fCPU,
	}
	fmt.Printf("[Success] Monitoring CPU of Pod: %s\n", podID)
}

func (c *Collector) RemovePod(podID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if t, ok := c.targets[podID]; ok {
		t.CPUFile.Close()
		delete(c.targets, podID)
		fmt.Printf("[Info] Pod removed from monitoring: %s\n", podID)
	}
}

func (c *Collector) Start() {
	go c.loopCPU()
}

func (c *Collector) loopCPU() {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for range ticker.C {
		c.collectCPU()
	}
}
