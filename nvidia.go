package main

import (
	"context"
	"fmt"
	"net/http"

	"cloud.google.com/go/compute/metadata"
	"github.com/NVIDIA/gpu-monitoring-tools/bindings/go/nvml"
	"golang.org/x/sync/errgroup"
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
