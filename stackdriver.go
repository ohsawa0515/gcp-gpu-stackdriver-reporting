package main

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/compute/metadata"
	monitoring "cloud.google.com/go/monitoring/apiv3"
	googlepb "github.com/golang/protobuf/ptypes/timestamp"
	metricpb "google.golang.org/genproto/googleapis/api/metric"
	monitoredrespb "google.golang.org/genproto/googleapis/api/monitoredres"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

const METRIC_TYPE_PREFIX = "custom.googleapis.com/"

type gpuStackdriverClient struct {
	projectID    string
	instanceID   string
	instanceName string
	zone         string
	ctx          context.Context
	metricClient *monitoring.MetricClient
}

func NewGpuStackdriverClient(ctx context.Context, meClient *metadata.Client) (*gpuStackdriverClient, error) {
	moClient, err := monitoring.NewMetricClient(ctx)
	if err != nil {
		return nil, err
	}
	projectID, err := meClient.ProjectID()
	if err != nil {
		return nil, fmt.Errorf("failed to get gce metadata project_id: %v", err)
	}
	instanceID, err := meClient.InstanceID()
	if err != nil {
		return nil, fmt.Errorf("failed to get gce metadata instance_id: %v", err)
	}
	instanceName, err := meClient.InstanceName()
	if err != nil {
		return nil, fmt.Errorf("failed to get gce metadata instance_name: %v", err)
	}
	zone, err := meClient.Zone()
	if err != nil {
		return nil, fmt.Errorf("failed to get gce metadata zone: %v", err)
	}

	return &gpuStackdriverClient{
		projectID:    projectID,
		instanceID:   instanceID,
		instanceName: instanceName,
		zone:         zone,
		ctx:          ctx,
		metricClient: moClient,
	}, nil
}

func (client *gpuStackdriverClient) reportGpuMetric(metricType string, value float64) error {
	// Prepares an individual data point
	dataPoint := &monitoringpb.Point{
		Interval: &monitoringpb.TimeInterval{
			EndTime: &googlepb.Timestamp{
				Seconds: time.Now().Unix(),
			},
		},
		Value: &monitoringpb.TypedValue{
			Value: &monitoringpb.TypedValue_DoubleValue{
				DoubleValue: value,
			},
		},
	}

	// Writes time series data.
	if err := client.metricClient.CreateTimeSeries(client.ctx, &monitoringpb.CreateTimeSeriesRequest{
		Name: fmt.Sprintf("projects/%s", client.projectID),
		TimeSeries: []*monitoringpb.TimeSeries{
			{
				Metric: &metricpb.Metric{
					Type: METRIC_TYPE_PREFIX + metricType,
					Labels: map[string]string{
						"instance_name": client.instanceName,
					},
				},
				Resource: &monitoredrespb.MonitoredResource{
					Type: "gce_instance",
					Labels: map[string]string{
						"project_id":  client.projectID,
						"instance_id": client.instanceID,
						"zone":        client.zone,
					},
				},
				Points: []*monitoringpb.Point{
					dataPoint,
				},
			},
		},
	}); err != nil {
		return fmt.Errorf("failed to write time series data: %v", err)
	}

	return nil
}
