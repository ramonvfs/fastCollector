package collector

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// readUsageUsec reads the usage_usec value from the CPU file of a PodTarget
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
	return 0, fmt.Errorf("usage_usec not found")
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
		// Convert to millicores:
		// Delta in us / 10ms (10000us) * 1000 (conversion to millicores)
		// Simplified: Delta / 10
		miliCore := float64(totalDelta) / 10.0

		now := time.Now().UnixMilli()

		//timestamp, cpu_millicore, num_rx
		line := fmt.Sprintf("%d,%.2f\n", now, miliCore)
		c.LogFile.WriteString(line)
	}
}
