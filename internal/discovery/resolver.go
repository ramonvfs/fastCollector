package discovery

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ResolveCPUPath(podUID string) (string, error) {
	uidSafe := strings.ReplaceAll(podUID, "-", "_")

	bases := []string{
		"/sys/fs/cgroup/kubepods.slice/kubepods-pod%s.slice/cpu.stat",
		"/sys/fs/cgroup/kubepods.slice/kubepods-burstable.slice/kubepods-burstable-pod%s.slice/cpu.stat",
		"/sys/fs/cgroup/kubepods.slice/kubepods-besteffort.slice/kubepods-besteffort-pod%s.slice/cpu.stat",
	}

	for _, base := range bases {
		path := fmt.Sprintf(base, uidSafe)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("CPU path not found for pod %s", podUID)
}

func ResolveNetPID(podUID string) (string, error) {
	uidSafe := strings.ReplaceAll(podUID, "-", "_")

	bases := []string{
		"/sys/fs/cgroup/kubepods.slice/kubepods-pod%s.slice/",
		"/sys/fs/cgroup/kubepods.slice/kubepods-burstable.slice/kubepods-burstable-pod%s.slice/",
		"/sys/fs/cgroup/kubepods.slice/kubepods-besteffort.slice/kubepods-besteffort-pod%s.slice/",
	}

	for _, base := range bases {
		podPath := fmt.Sprintf(base, uidSafe)

		if _, err := os.Stat(podPath); err == nil {
			entries, err := os.ReadDir(podPath)
			if err != nil {
				continue
			}

			for _, entry := range entries {
				if entry.IsDir() && strings.Contains(entry.Name(), ".scope") {
					pidFilePath := filepath.Join(podPath, entry.Name(), "cgroup.procs")
					content, err := os.ReadFile(pidFilePath)
					if err == nil {
						pids := strings.Fields(string(content))
						if len(pids) > 0 {
							return pids[0], nil
						}
					}
				}
			}
		}
	}

	return "", fmt.Errorf("PID not found for pod %s", podUID)
}

func ResolveNetPath(podUID string) (string, error) {
	pid, err := ResolveNetPID(podUID)
	if err != nil {
		return "", err
	}

	netPath := fmt.Sprintf("/host/proc/%s/net/dev", pid)

	if _, err := os.Stat(netPath); err != nil {
		return "", fmt.Errorf("Network path %s not accessible: %v", netPath, err)
	}

	return netPath, nil
}
