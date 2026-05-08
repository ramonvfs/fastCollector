package collector

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
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

		if t.LastCPUUsage > 0 {
			totalDelta += (atual - t.LastCPUUsage)
		}

		t.LastCPUUsage = atual
		numPods++
	}

	// Calcula a média do nó para este instante de 10ms
	if numPods > 0 {
		media := float64(totalDelta) / float64(numPods)
		c.CPUBuffer = append(c.CPUBuffer, MetricPoint{
			Value: media,
		})
	}
}
