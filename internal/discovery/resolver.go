package discovery

import (
	"fmt"
	"os"
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
