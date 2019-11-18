package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cloud.google.com/go/compute/metadata"
	"github.com/NVIDIA/gpu-monitoring-tools/bindings/go/nvml"
	"golang.org/x/sync/errgroup"
)

/*
Number of seconds between sending metrics to Stackdriver
*/
const SendIntervalSecond = 60

/*
The number of seconds between collecting GPU metrics.
Send the average value to Stackdriver at SendIntervalSecond.
*/
const CollectIntervalSecond = 10

func gpuUtilizationTicker(ctx context.Context, client *gpuStackdriverClient, devices []*nvml.Device) error {
	sendTicker := time.NewTicker(time.Second * SendIntervalSecond)
	collectTicker := time.NewTicker(time.Second * CollectIntervalSecond)
	defer sendTicker.Stop()
	defer collectTicker.Stop()

	metric := float64(0)
	count := 0
	for {
		select {
		case <-sendTicker.C:
			log.Println("Send GPU utilization")
			sentMetric := metric / float64(count)
			log.Println(sentMetric)
			if err := client.reportGpuMetric("gpu_utilization", float64(sentMetric)); err != nil {
				return err
			}
			metric = 0
			count = 0
		case <-collectTicker.C:
			log.Println("Collect GPU utilization")
			for i, device := range devices {
				st, err := device.Status()
				if err != nil {
					return fmt.Errorf("error getting device %d status: %v\n", i, err)
				}
				metric += float64(*st.Utilization.GPU)
				count += 1
			}
			log.Println(metric)
			log.Println(count)
		case <-ctx.Done():
			log.Println("Stop GPU utilization")
			return ctx.Err()
		}
	}
}

func gpuMemoryUtilizationTicker(ctx context.Context, client *gpuStackdriverClient, devices []*nvml.Device) error {
	sendTicker := time.NewTicker(time.Second * SendIntervalSecond)
	collectTicker := time.NewTicker(time.Second * CollectIntervalSecond)
	defer sendTicker.Stop()
	defer collectTicker.Stop()

	metric := float64(0)
	count := 0
	for {
		select {
		case <-sendTicker.C:
			sentMetric := metric / float64(count)
			if err := client.reportGpuMetric("gpu_memory_utilization", float64(sentMetric)); err != nil {
				return err
			}
			metric = 0
			count = 0
		case <-collectTicker.C:
			for i, device := range devices {
				st, err := device.Status()
				if err != nil {
					return fmt.Errorf("error getting device %d status: %v\n", i, err)
				}
				metric += float64(*st.Utilization.Memory)
				count += 1
			}
		case <-ctx.Done():
			log.Println("Stop GPU memory utilization ticker")
			return ctx.Err()
		}
	}
}

func gpuTemperatureTicker(ctx context.Context, client *gpuStackdriverClient, devices []*nvml.Device) error {
	sendTicker := time.NewTicker(time.Second * SendIntervalSecond)
	collectTicker := time.NewTicker(time.Second * CollectIntervalSecond)
	defer sendTicker.Stop()
	defer collectTicker.Stop()

	metric := float64(0)
	count := 0
	for {
		select {
		case <-sendTicker.C:
			sentMetric := metric / float64(count)
			if err := client.reportGpuMetric("gpu_temperature", float64(sentMetric)); err != nil {
				return err
			}
		case <-collectTicker.C:
			for i, device := range devices {
				st, err := device.Status()
				if err != nil {
					return fmt.Errorf("error getting device %d status: %v\n", i, err)
				}
				metric += float64(*st.Temperature)
				count += 1
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

func Run() error {
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

	parent := signalContext(ctx)
	eg, child := errgroup.WithContext(parent)

	eg.Go(func() error {
		return gpuUtilizationTicker(child, client, devices)
	})

	eg.Go(func() error {
		return gpuMemoryUtilizationTicker(child, client, devices)
	})

	eg.Go(func() error {
		return gpuTemperatureTicker(child, client, devices)
	})

	return eg.Wait()
}
