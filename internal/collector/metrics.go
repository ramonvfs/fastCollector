package collector

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
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

func (c *Collector) readRxPackets(target *PodTarget) (uint64, error) {
	_, err := target.NetFile.Seek(0, 0)
	if err != nil {
		return 0, fmt.Errorf("Error to reset netfile: %v", err)
	}
	scanner := bufio.NewScanner(target.NetFile)

	const ifaceToFind = "eth0"

	for scanner.Scan() {
		line := scanner.Text()

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		iface := strings.TrimSpace(parts[0])
		if iface != ifaceToFind && !strings.HasPrefix(iface, ifaceToFind+"@") {
			continue
		}

		cols := strings.Fields(strings.TrimSpace(parts[1]))
		if len(cols) >= 2 {
			return strconv.ParseUint(cols[1], 10, 64)
		}
	}

	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("Error scanning net file: %v", err)
	}
	return 0, fmt.Errorf("Interface %s not found in %s", ifaceToFind, target.NetFile.Name())
}

func (c *Collector) collectCPU() float64 {
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
		return miliCore
	}
	return 0.0
}

func (c *Collector) collectNet() float64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	var totalRx uint64
	numPods := 0

	for _, t := range c.targets {
		rx, err := c.readRxPackets(t)
		if err != nil {
			continue
		}

		if t.LastNetRx > 0 && rx > t.LastNetRx {
			totalRx += (rx - t.LastNetRx)
		}
		t.LastNetRx = rx
		numPods++
	}

	if numPods > 0 {
		// Convert to RPM (100ms interval, so multiply by 600 to get per minute):
		netRPM := float64(totalRx) / float64(numPods) * 600.0
		return netRPM
	}
	return 0.0
}
