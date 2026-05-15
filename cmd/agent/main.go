package main

import (
	"fastcollector/internal/collector"
	"fastcollector/internal/discovery"
	"fmt"
	"os"
)

func main() {
	if os.Getenv("APP_LABEL") == "" {
		fmt.Println("[Error] The APP_LABEL environment variable is required. Example: app=my-app")
		os.Exit(1)
	}

	col, err := collector.NewCollector("/var/log/ia-data/metrics.csv")
	if err != nil {
		fmt.Printf("[Error] Failed to create collector: %v\n", err)
		os.Exit(1)
	}
	col.Start()

	watcher, err := discovery.NewPodWatcher(col)
	if err != nil {
		fmt.Printf("[Error] Failed to start K8s Watcher: %v\n", err)
		os.Exit(1)
	}
	watcher.Start()

	select {}
}
