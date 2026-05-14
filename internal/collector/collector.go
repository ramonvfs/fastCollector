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
	LastCPUUsage uint64
	LastNetRx    uint64
	CPUFile      *os.File
	NetFile      *os.File
}

type Collector struct {
	mu           sync.RWMutex
	targets      map[string]*PodTarget
	LogFile      *os.File
	lastNetValue float64
}

func NewCollector(logPath string) (*Collector, error) {
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	info, _ := f.Stat()
	if info.Size() == 0 {
		f.WriteString("timestamp,cpu_millicore,net_rpm\n")
	}

	return &Collector{
		targets: make(map[string]*PodTarget),
		LogFile: f,
	}, nil
}

func (c *Collector) AddPod(podID, cpuPath, netPath string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	fCPU, err := os.Open(cpuPath)
	if err != nil {
		fmt.Printf("[Error] Failed to open CPUFile %s for Pod %s: %v\n", cpuPath, podID, err)
		return
	}

	fNet, err := os.Open(netPath)
	if err != nil {
		fmt.Printf("[Error] Failed to open NetFile %s for Pod %s: %v\n", netPath, podID, err)
		fCPU.Close()
		return
	}

	c.targets[podID] = &PodTarget{
		ID:      podID,
		CPUFile: fCPU,
		NetFile: fNet,
	}
	fmt.Printf("[Success] Monitoring CPU and Network of Pod: %s\n", podID)
}

func (c *Collector) RemovePod(podID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if t, ok := c.targets[podID]; ok {
		t.CPUFile.Close()
		t.NetFile.Close()
		delete(c.targets, podID)
		fmt.Printf("[Info] Pod removed from monitoring: %s\n", podID)
	}
}

func (c *Collector) Start() {
	go c.loopCollect()
}

func (c *Collector) loopCollect() {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	counter := 0

	for range ticker.C {
		cpuValue := c.collectCPU()

		if counter >= 10 {
			c.lastNetValue = c.collectNet()
			counter = 0
		}

		now := time.Now().UnixMilli()

		//timestamp, cpu_millicore, net_rpm
		line := fmt.Sprintf("%d,%.2f,%.2f\n", now, cpuValue, c.lastNetValue)
		_, err := c.LogFile.WriteString(line)
		if err != nil {
			fmt.Printf("Error to write to log: %v\n", err)
		}

		counter++
	}

}
