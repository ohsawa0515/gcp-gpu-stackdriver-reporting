package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/NVIDIA/gpu-monitoring-tools/bindings/go/nvml"
)

const INTERNAL_SECOND = 60

func gpuUtilizationTicker(ctx context.Context, client *gpuStackdriverClient, devices []*nvml.Device) error {
	ticker := time.NewTicker(time.Second * INTERNAL_SECOND)
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
			}
		case <-ctx.Done():
			log.Println("Stop GPU utilization ticker")
			return ctx.Err()
		}
	}
}

func gpuMemoryUtilizationTicker(ctx context.Context, client *gpuStackdriverClient, devices []*nvml.Device) error {
	ticker := time.NewTicker(time.Second * INTERNAL_SECOND)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			for i, device := range devices {
				st, err := device.Status()
				if err != nil {
					return fmt.Errorf("Error getting device %d status: %v\n", i, err)
				}
				if err := client.reportGpuMetric("gpu_memory_utilization", float64(*st.Utilization.Memory)); err != nil {
					return err
				}
			}
		case <-ctx.Done():
			log.Println("Stop GPU memory utilization ticker")
			return ctx.Err()
		}
	}
}

func gpuTemperatureTicker(ctx context.Context, client *gpuStackdriverClient, devices []*nvml.Device) error {
	ticker := time.NewTicker(time.Second * INTERNAL_SECOND)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			for i, device := range devices {
				st, err := device.Status()
				if err != nil {
					return fmt.Errorf("Error getting device %d status: %v\n", i, err)
				}
				if err := client.reportGpuMetric("gpu_temperature", float64(*st.Temperature)); err != nil {
					return err
				}
			}
		case <-ctx.Done():
			log.Println("Stop GPU temperature ticker")
			return ctx.Err()
		}
	}
}

func signalContext(ctx context.Context) context.Context {
	parent, cancelParent := context.WithCancel(ctx)

	go func() {
		defer cancelParent()

		sig := make(chan os.Signal, 1)
		signal.Notify(sig,
			syscall.SIGHUP,
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGQUIT)
		defer signal.Stop(sig)

		select {
		case <-parent.Done():
			log.Println("Cancel from parent")
			return
		case <-sig:
			log.Println("Cancel from signal")
			return
		}
	}()

	return parent
}

func main() {
	if err := nvidia(); err != nil {
		log.Fatal(err)
	}
}
