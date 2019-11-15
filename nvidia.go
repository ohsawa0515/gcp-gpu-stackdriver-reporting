package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cloud.google.com/go/compute/metadata"
	"github.com/NVIDIA/gpu-monitoring-tools/bindings/go/nvml"
)

func nvidia() error {
	nvml.Init()
	defer nvml.Shutdown()

	count, err := nvml.GetDeviceCount()
	if err != nil {
		return fmt.Errorf("error getting device count: %v", err)
	}

	var devices []*nvml.Device
	for i := uint(0); i < count; i++ {
		device, err := nvml.NewDevice(i)
		if err != nil {
			return fmt.Errorf("Error getting device %d: %v\n", i, err)
		}
		devices = append(devices, device)
	}

	ctx := context.Background()

	gce := metadata.NewClient(&http.Client{})
	client, err := NewGpuStackdriverClient(ctx, gce)
	if err != nil {
		return err
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(time.Second * 60)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			for i, device := range devices {
				st, err := device.Status()
				if err != nil {
					return fmt.Errorf("Error getting device %d status: %v\n", i, err)
				}
				if err := client.reportGpuMetric("gpu_utilization", float64(*st.Utilization.GPU)); err != nil {
					return err
				}
				if err := client.reportGpuMetric("gpu_memory_utilization", float64(*st.Utilization.Memory)); err != nil {
					return err
				}
				if err := client.reportGpuMetric("gpu_temperature", float64(*st.Temperature)); err != nil {
					return err
				}
			}
		case <-sigs:
			return nil
		}
	}
}
