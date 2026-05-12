package collector

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func (c *Collector) readUsageUsec(target *PodTarget) (uint64, error) {
	target.CPUFile.Seek(0, 0)
	scanner := bufio.NewScanner(target.CPUFile)

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "usage_usec") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return strconv.ParseUint(parts[1], 10, 64)
			}
		}
	}
	return 0, fmt.Errorf("usage_usec não encontrado")
}

func (c *Collector) collectCPU() {
	c.mu.Lock()
	defer c.mu.Unlock()

	var totalDelta uint64
	numPods := 0

	for _, t := range c.targets {
		atual, err := c.readUsageUsec(t)
		if err != nil {
			continue
		}

		if t.LastCPUUsage > 0 && atual > t.LastCPUUsage {
			totalDelta += (atual - t.LastCPUUsage)
		}
		t.LastCPUUsage = atual
		numPods++
	}

	if numPods > 0 {
		// CONVERSÃO PARA MILLICORE:
		// Delta em us / 10ms (10000us) * 1000 (conversão millicore)
		// Simplificado: Delta / 10
		miliCore := float64(totalDelta) / 10.0

		now := time.Now().UnixMilli()

		// 1. Salva no buffer de memória
		c.CPUBuffer = append(c.CPUBuffer, MetricPoint{
			Timestamp: now,
			Value:     miliCore,
		})

		// 2. Salva direto no arquivo para o Dataset da IA
		// Formato: timestamp, valor
		line := fmt.Sprintf("%d,%.2f\n", now, miliCore)
		c.LogFile.WriteString(line)
	}
}
